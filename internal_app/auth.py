from functools import wraps
from flask import request, abort, current_app

# --- Simulated Development User Configuration ---
# These values are used by the require_authenticated decorator when
# FLASK_ENV is 'development' and no actual Remote-User header is present.
DEV_REMOTE_USER = 'dev_user'
DEV_REMOTE_NAME = 'Development User (Simulated)'
DEV_REMOTE_GROUPS = 'users,sudoers'  # Ensure 'sudoers' for testing admin functionality
DEV_REMOTE_EMAIL = 'dev_user@example.com (Simulated)'
# --- End Simulated Development User Configuration ---

def _simulate_dev_headers_if_needed(request_obj, app_obj):
    """
    Internal helper to simulate development headers if FLASK_ENV is 'development'
    and essential headers are missing from the request environment.
    Modifies request_obj.environ in place.
    """
    if app_obj.config.get('ENV') == 'development':
        if 'HTTP_REMOTE_USER' not in request_obj.environ:
            app_obj.logger.info(
                f"DEV MODE: Simulating headers for {request_obj.path} using predefined development user."
            )
            request_obj.environ['HTTP_REMOTE_USER'] = DEV_REMOTE_USER
            request_obj.environ['HTTP_REMOTE_NAME'] = DEV_REMOTE_NAME
            request_obj.environ['HTTP_REMOTE_GROUPS'] = DEV_REMOTE_GROUPS
            request_obj.environ['HTTP_REMOTE_EMAIL'] = DEV_REMOTE_EMAIL
            # Flask's request.headers will pick up these changes from environ.

def get_user_groups():
    """Helper function to get user groups from request headers."""
    remote_groups_header = request.headers.get('Remote-Groups')
    if remote_groups_header:
        return [g.strip() for g in remote_groups_header.split(',')]
    return []

def is_sudoer():
    """Checks if the current user is part of the 'sudoers' group."""
    return 'sudoers' in get_user_groups()

def require_sudoer(f):
    """
    Decorator to ensure the user is part of the 'sudoers' group.
    Aborts with 403 Forbidden if not.
    """
    @wraps(f)
    def decorated_function(*args, **kwargs):
        _simulate_dev_headers_if_needed(request, current_app) # Ensure headers are present/simulated

        if not request.headers.get('Remote-User'): # This check is now more robust
            current_app.logger.warning(f"Access denied to sudoer-only route {request.path}: No Remote-User header after potential simulation.")
            abort(401) # Should ideally not be hit if simulation works, but good safeguard.
        
        if not is_sudoer():
            current_app.logger.warning(
                f"Forbidden access attempt to {request.path} by user {request.headers.get('Remote-User', 'Unknown')}. "
                f"User is not in 'sudoers' group. Groups: {request.headers.get('Remote-Groups', 'None')}."
            )
            # We'll need a 403.html template later
            abort(403, description="You do not have sufficient permissions to access this resource. 'sudoers' group membership is required.")
        return f(*args, **kwargs)
    return decorated_function

def require_authenticated(f):
    """
    Decorator to ensure the user is authenticated (Remote-User header is present).
    Aborts with 401 Unauthorized if not.
    Simulates headers in development if FLASK_ENV is 'development' and Remote-User is missing,
    using the globally defined DEV_* constants.
    """
    @wraps(f)
    def decorated_function(*args, **kwargs):
        _simulate_dev_headers_if_needed(request, current_app) # Ensure headers are present/simulated
        
        # Now perform the actual authentication check using request.headers
        if not request.headers.get('Remote-User'): # This check is now more robust
            current_app.logger.info(
                f"Authentication failed for {request.path}: No Remote-User header present (after potential dev simulation). IP: {request.remote_addr}"
            )
            abort(401)
        
        return f(*args, **kwargs)
    return decorated_function
