from flask import render_template, request, current_app, abort, redirect, url_for, flash
from . import internal_bp
from .models import db, User, RadUserGroup, RadCheck, normalize_mac, format_mac_display, DeviceView
from .auth import require_authenticated, require_sudoer, is_sudoer as check_is_sudoer
import os
import re # For MAC address validation
from sqlalchemy.exc import IntegrityError, SQLAlchemyError
from sqlalchemy.orm import aliased

# --- Configuration for Device Management ---
VALID_VLAN_NAMES = ["trusted", "iot", "guest"]
# Relax regex slightly to allow various separators for input, but we'll normalize it.
# Using the DB format (12 hex chars) for internal consistency after normalization.
MAC_ADDRESS_REGEX_INPUT = re.compile(r"^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$")
# --- End Configuration ---

@internal_bp.route('/')
@require_authenticated
def landing_page():
    """
    Serves the main landing page for the /intern section.
    This page will include navigation to sub-tools and display user information.
    The `require_authenticated` decorator handles simulating headers in development.
    """
    # Headers are now guaranteed to be present (either real or simulated by decorator)
    remote_user = request.headers.get('Remote-User')
    remote_name = request.headers.get('Remote-Name')
    remote_groups = request.headers.get('Remote-Groups')
    remote_email = request.headers.get('Remote-Email')
    
    current_is_sudoer = check_is_sudoer() # Uses the (potentially simulated) headers

    nav_items = [
        # The 'url' key here should be the endpoint name for url_for(), not the path itself.
        # The url_for() in the template will correctly build the path.
        {'name': 'Device Management', 'url': 'internal_bp.devices_list', 'requires_sudoer': False}
        # Add more tools here
    ]
    
    # accessible_nav_items = [item for item in nav_items if not item.get('requires_sudoer') or current_is_sudoer]

    return render_template('internal_app/landing.html',
                           remote_user=remote_user,
                           remote_name=remote_name,
                           remote_groups=remote_groups,
                           remote_email=remote_email,
                           is_sudoer=current_is_sudoer, # Pass sudoer status to template
                           nav_items=nav_items)


# Device Management Routes
@internal_bp.route('/devices')
@require_authenticated
def devices_list():
    device_views = []
    try:
        # Query users and join with radusergroup to get vlan_name (groupname)
        # Using aliased for RadUserGroup in case of future more complex joins
        RUG = aliased(RadUserGroup)
        query = db.session.query(
            User.username, # This is the normalized MAC
            User.description,
            RUG.groupname.label('vlan_name')
        ).outerjoin(RUG, User.username == RUG.username).order_by(User.username)
        
        results = query.all()

        for row in results:
            # Create DeviceView using the normalized MAC (username) from DB,
            # the DeviceView will handle the display formatting internally.
            device_views.append(DeviceView(mac_address=row.username, 
                                           description=row.description,
                                           vlan_name=row.vlan_name))
    except SQLAlchemyError as e:
        current_app.logger.error(f"Database error fetching devices: {e}")
        flash("Error fetching devices from the database.", "danger")
    
    return render_template('internal_app/devices.html', devices=device_views, is_sudoer=check_is_sudoer())

@internal_bp.route('/devices/add', methods=['GET', 'POST'])
@require_sudoer
def add_device():
    if request.method == 'POST':
        mac_address_input = request.form.get('mac_address', '').strip()
        vlan_name = request.form.get('vlan_name')
        description = request.form.get('description', '').strip()
        error = False

        if not mac_address_input:
            flash('MAC Address is required.', 'danger')
            error = True
        # Use the regex specifically for input validation
        elif not MAC_ADDRESS_REGEX_INPUT.match(mac_address_input):
            flash('Invalid MAC Address format. Use XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX.', 'danger')
            error = True
        
        if not vlan_name:
            flash('VLAN Name is required.', 'danger')
            error = True
        elif vlan_name not in VALID_VLAN_NAMES:
            flash(f"Invalid VLAN Name. Must be one of: {', '.join(VALID_VLAN_NAMES)}.", 'danger')
            error = True

        # Create a DeviceView using the potentially unnormalized input for form repopulation
        temp_device_view = DeviceView(mac_address=mac_address_input, description=description, vlan_name=vlan_name)

        if error:
            return render_template('internal_app/device_form.html', 
                                   device_view=temp_device_view,
                                   valid_vlan_names=VALID_VLAN_NAMES,
                                   form_action_url=url_for('internal_bp.add_device'))

        # Normalize MAC for database operations
        normalized_mac = normalize_mac(mac_address_input)
        
        # This check is still good, even after regex, to ensure DB uniqueness for normalized form
        existing_user = User.query.filter_by(username=normalized_mac).first()
        if existing_user:
            # Use the formatted_mac property for the flash message
            flash(f'Device with MAC address {temp_device_view.formatted_mac} already exists.', 'danger')
            return render_template('internal_app/device_form.html',
                                   device_view=temp_device_view,
                                   valid_vlan_names=VALID_VLAN_NAMES,
                                   form_action_url=url_for('internal_bp.add_device'))
        
        try:
            # Start transaction
            new_user = User(username=normalized_mac, description=description)
            db.session.add(new_user)

            new_rad_user_group = RadUserGroup(username=normalized_mac, groupname=vlan_name, priority=0)
            db.session.add(new_rad_user_group)

            new_rad_check = RadCheck(username=normalized_mac, attribute='Cleartext-Password', op=':=', value=normalized_mac)
            db.session.add(new_rad_check)
            
            db.session.commit()
            # Use the formatted_mac property for the flash message
            flash(f'Device {temp_device_view.formatted_mac} added successfully!', 'success')
            current_app.logger.info(f"Device {temp_device_view.formatted_mac} (normalized: {normalized_mac}) added by {request.headers.get('Remote-User')}")
            return redirect(url_for('internal_bp.devices_list'))
        except SQLAlchemyError as e:
            db.session.rollback()
            flash('Database error occurred while adding device.', 'danger')
            current_app.logger.error(f"SQLAlchemyError adding device {mac_address_input}: {e}")
            # Use the formatted_mac property for form repopulation after error
            # Recreate temp_device_view in case rollback affected it
            temp_device_view = DeviceView(mac_address=mac_address_input, description=description, vlan_name=vlan_name)
            return render_template('internal_app/device_form.html', 
                                   device_view=temp_device_view,
                                   valid_vlan_names=VALID_VLAN_NAMES,
                                   form_action_url=url_for('internal_bp.add_device'))

    return render_template('internal_app/device_form.html', 
                           device_view=None, 
                           valid_vlan_names=VALID_VLAN_NAMES,
                           form_action_url=url_for('internal_bp.add_device'))

@internal_bp.route('/devices/edit/<mac_address_display>', methods=['GET', 'POST'])
@require_sudoer
def edit_device(mac_address_display):
    # mac_address_display here is the original format from the URL
    normalized_mac = normalize_mac(mac_address_display)
    # Query using the normalized MAC stored in the database
    user_to_edit = User.query.filter_by(username=normalized_mac).first_or_404()
    rad_group_to_edit = RadUserGroup.query.filter_by(username=normalized_mac).first()

    # Create a DeviceView for displaying current data in the form
    # Use the original mac_address_display from the URL for the form input field
    current_device_view = DeviceView(mac_address=mac_address_display, 
                                     description=user_to_edit.description,
                                     vlan_name=rad_group_to_edit.groupname if rad_group_to_edit else None)

    if request.method == 'POST':
        new_vlan_name = request.form.get('vlan_name')
        new_description = request.form.get('description', '').strip()
        error = False

        if not new_vlan_name:
            flash('VLAN Name is required.', 'danger')
            error = True
        elif new_vlan_name not in VALID_VLAN_NAMES:
            flash(f"Invalid VLAN Name. Must be one of: {', '.join(VALID_VLAN_NAMES)}.", 'danger')
            error = True
        
        # MAC address is not editable in this form, so no validation needed for the input field's value here.
        # The mac_address_display from the URL is used to identify the device.
        
        if error:
            # Rebuild DeviceView for form repopulation with current (pre-edit attempt) data
            # Use the original mac_address_display from the URL, but updated description/vlan from form attempt
            temp_device_view_on_error = DeviceView(mac_address=mac_address_display, 
                                                   description=new_description, 
                                                   vlan_name=new_vlan_name)
            return render_template('internal_app/device_form.html', 
                                   device_view=temp_device_view_on_error,
                                   valid_vlan_names=VALID_VLAN_NAMES,
                                   form_action_url=url_for('internal_bp.edit_device', mac_address_display=mac_address_display))
        
        try:
            user_to_edit.description = new_description
            if rad_group_to_edit:
                rad_group_to_edit.groupname = new_vlan_name
            else: # Should not happen if data is consistent from add, but handle defensively
                new_rad_group = RadUserGroup(username=normalized_mac, groupname=new_vlan_name, priority=0)
                db.session.add(new_rad_group)
            
            # RadCheck usually doesn't change on edit unless MAC itself changes (which it doesn't here)
            
            db.session.commit()
            # Use the formatted_mac property for the flash message. 
            # Need a DeviceView object to access formatted_mac. Create one using the original display MAC.
            flash_device_view = DeviceView(mac_address=mac_address_display, description=new_description, vlan_name=new_vlan_name)
            flash(f'Device {flash_device_view.formatted_mac} updated successfully!', 'success')
            current_app.logger.info(f"Device {flash_device_view.formatted_mac} updated by {request.headers.get('Remote-User')}")
            return redirect(url_for('internal_bp.devices_list'))
        except SQLAlchemyError as e:
            db.session.rollback()
            flash('Database error occurred while updating device.', 'danger')
            current_app.logger.error(f"SQLAlchemyError updating device {mac_address_display}: {e}")
            # Rebuild DeviceView for form repopulation with updated data
            temp_device_view_on_error = DeviceView(mac_address=mac_address_display, 
                                                   description=new_description, 
                                                   vlan_name=new_vlan_name)
            return render_template('internal_app/device_form.html', 
                                   device_view=temp_device_view_on_error,
                                   valid_vlan_names=VALID_VLAN_NAMES,
                                   form_action_url=url_for('internal_bp.edit_device', mac_address_display=mac_address_display))

    # For GET request, render the form with existing data
    return render_template('internal_app/device_form.html', 
                           device_view=current_device_view, 
                           valid_vlan_names=VALID_VLAN_NAMES,
                           form_action_url=url_for('internal_bp.edit_device', mac_address_display=mac_address_display))

@internal_bp.route('/devices/delete/<mac_address_display>', methods=['POST'])
@require_sudoer
def delete_device(mac_address_display):
    # mac_address_display here is the original format from the URL
    normalized_mac = normalize_mac(mac_address_display)
    
    # Use a DeviceView to get the formatted MAC for the flash message
    # We need to create a temporary one as we are about to delete the data
    flash_device_view = DeviceView(mac_address=mac_address_display, description=None, vlan_name=None)

    try:
        # Find the related entries using the normalized MAC
        user_to_delete = User.query.filter_by(username=normalized_mac).first_or_404()
        rad_group_to_delete = RadUserGroup.query.filter_by(username=normalized_mac).first()
        rad_check_to_delete = RadCheck.query.filter_by(username=normalized_mac).first()

        # Delete the entries
        if rad_check_to_delete:
            db.session.delete(rad_check_to_delete)
        if rad_group_to_delete:
            db.session.delete(rad_group_to_delete)
        # Always delete the user entry as it's the primary record
        db.session.delete(user_to_delete)
        
        db.session.commit()
        # Use the formatted_mac property for the flash message
        flash(f'Device {flash_device_view.formatted_mac} deleted successfully.', 'success')
        current_app.logger.info(f"Device {flash_device_view.formatted_mac} (normalized: {normalized_mac}) deleted by {request.headers.get('Remote-User')}")
    except SQLAlchemyError as e:
        db.session.rollback()
        flash(f'Database error occurred while deleting device {flash_device_view.formatted_mac}.', 'danger')
        current_app.logger.error(f"SQLAlchemyError deleting device {mac_address_display}: {e}")
    except Exception as e:
        flash(f'An unexpected error occurred while deleting device {flash_device_view.formatted_mac}.', 'danger')
        current_app.logger.error(f"Unexpected error deleting device {mac_address_display}: {e}")
        
    return redirect(url_for('internal_bp.devices_list'))
