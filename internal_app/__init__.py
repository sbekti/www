from flask import Blueprint

internal_bp = Blueprint(
    'internal_bp', __name__,
    template_folder='../templates/internal_app', # Points to www/templates/internal_app
    static_folder='../static', # Can share the main static folder or have its own
    url_prefix='/intern' # All routes in this blueprint will be prefixed with /intern
)

# Import routes after blueprint definition to avoid circular imports
from . import routes, models # models might not be directly used here but good to include for structure
