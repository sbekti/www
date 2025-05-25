from flask import Blueprint
from .routes import register_internal_routes # Import the new function

# Blueprint for subdomain access (e.g., intern.corp.bekti.com)
internal_subdomain_bp = Blueprint(
    'internal_subdomain', __name__,
    template_folder='../templates/internal_app', # Points to www/templates/internal_app
    static_folder='../static' # Can share the main static folder or have its own
    # No url_prefix here; it will be at the root of the subdomain
)

# Blueprint for path-based access (e.g., /intern for development)
internal_path_bp = Blueprint(
    'internal_path', __name__,
    template_folder='../templates/internal_app',
    static_folder='../static',
    url_prefix='/intern' # All routes in this blueprint will be prefixed with /intern
)

# Register the defined routes onto both blueprints
register_internal_routes(internal_subdomain_bp)
register_internal_routes(internal_path_bp)

# Import models, typically done after blueprint and route setup if models need app context or db instance
# For now, keeping it similar to original structure.
from . import models # models might not be directly used here but good to include for structure
