from flask_sqlalchemy import SQLAlchemy
from sqlalchemy import text # For server_default with nextval

# Initialize SQLAlchemy. This will be properly configured with the app in app.py
db = SQLAlchemy()

# Helper to normalize MAC address for database storage (lowercase, no separators)
def normalize_mac(mac_address_str):
    if not mac_address_str:
        return None
    return "".join(filter(str.isalnum, mac_address_str)).lower()

# Helper to format normalized MAC address for display (XX:XX:XX:XX:XX:XX)
def format_mac_display(normalized_mac_str):
    if not normalized_mac_str or len(normalized_mac_str) != 12:
        return normalized_mac_str # Return as is if not a valid normalized MAC
    # Insert colons every two characters
    return ':'.join(normalized_mac_str[i:i+2] for i in range(0, 12, 2))

class User(db.Model):
    __tablename__ = 'users'
    # Assuming 'id' is an auto-incrementing primary key managed by the sequence
    id = db.Column(db.Integer, primary_key=True, server_default=text("nextval('users_id_seq'::regclass)"))
    username = db.Column(db.Text, nullable=False, unique=True) # This will store the normalized MAC address
    description = db.Column(db.Text, nullable=True)

    def __repr__(self):
        return f"<User {self.username}>"

class RadUserGroup(db.Model):
    __tablename__ = 'radusergroup'
    # Assuming 'id' is an auto-incrementing primary key
    id = db.Column(db.Integer, primary_key=True, server_default=text("nextval('radusergroup_id_seq'::regclass)"))
    username = db.Column(db.Text, nullable=False, default='') # Normalized MAC address
    groupname = db.Column(db.Text, nullable=False, default='') # VLAN name: "trusted", "iot", "guest"
    priority = db.Column(db.Integer, nullable=False, default=0)

    # Consider adding a composite unique constraint if a user can only be in one group or one group with a specific priority
    # db.UniqueConstraint('username', 'groupname', name='uix_user_group')

    def __repr__(self):
        return f"<RadUserGroup {self.username} - {self.groupname}>"

class RadCheck(db.Model):
    __tablename__ = 'radcheck'
    # Assuming 'id' is an auto-incrementing primary key
    id = db.Column(db.Integer, primary_key=True, server_default=text("nextval('radcheck_id_seq'::regclass)"))
    username = db.Column(db.Text, nullable=False, default='') # Normalized MAC address
    attribute = db.Column(db.Text, nullable=False, default='') # Should be "Cleartext-Password"
    op = db.Column(db.String(2), nullable=False, default=':=')
    value = db.Column(db.Text, nullable=False, default='') # Should be the normalized MAC address

    def __repr__(self):
        return f"<RadCheck {self.username} {self.attribute}>"

# Combined device view (not a table, but for conceptual representation or complex queries)
# This is more of a DTO or a structure you'd build from queries
class DeviceView:
    def __init__(self, mac_address, description, vlan_name, user_id=None, radusergroup_id=None, radcheck_id=None):
        # Store the original input/display format
        self.mac_address_display = mac_address
        # Store the normalized format for DB operations
        self.mac_address_db = normalize_mac(mac_address)
        self.description = description
        self.vlan_name = vlan_name
        # Optional IDs if needed
        self.user_id = user_id
        self.radusergroup_id = radusergroup_id
        self.radcheck_id = radcheck_id

    # Add a property to get the MAC address in the consistent display format
    @property
    def formatted_mac(self):
        # Use the stored normalized MAC for formatting
        return format_mac_display(self.mac_address_db)
