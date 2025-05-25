from flask import Flask, render_template, request, redirect, url_for, abort
import datetime
import os
import logging
from flask_sqlalchemy import SQLAlchemy # Added
from internal_app.models import db # Added
from internal_app import internal_bp # Added

import urllib.parse # For safely quoting password/username if constructing URI

app = Flask(__name__)

# Configuration
# Database Configuration - Prioritize individual components, fallback to DATABASE_URL
db_user = os.environ.get('DB_USERNAME')
db_password_raw = os.environ.get('DB_PASSWORD') # Raw password
db_host = os.environ.get('DB_HOST')
db_port = os.environ.get('DB_PORT')
db_name = os.environ.get('DB_NAME')

if all([db_user, db_password_raw is not None, db_host, db_port, db_name]):
    # URL-encode username and password when constructing the URI string
    db_password_encoded = urllib.parse.quote_plus(db_password_raw)
    db_user_encoded = urllib.parse.quote_plus(db_user)
    app.config['SQLALCHEMY_DATABASE_URI'] = \
        f"postgresql://{db_user_encoded}:{db_password_encoded}@{db_host}:{db_port}/{db_name}"
    app.logger.info("Configuring database via individual DB_* environment variables.")
else:
    app.config['SQLALCHEMY_DATABASE_URI'] = \
        os.environ.get('DATABASE_URL', 'postgresql://default_user:default_password@default_host:5432/default_db')
    app.logger.info("Configuring database via DATABASE_URL environment variable or default.")
    if not os.environ.get('DATABASE_URL'):
        app.logger.warning("DATABASE_URL not set, and not all individual DB_* variables were found. Using default DB URI.")


app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False
app.config['ENV'] = os.environ.get('FLASK_ENV', 'production') # For dev header simulation
# --- IMPORTANT: SECRET_KEY ---
# For production, this MUST be a complex, random, and secret value, ideally from an environment variable.
# For development, a simple key is okay, but ensure it's changed for production.
app.config['SECRET_KEY'] = os.environ.get('SECRET_KEY', 'dev_secret_key_for_flask_sessions')
# --- IMPORTANT: END SECRET_KEY ---


# Initialize extensions
db.init_app(app)

# Register Blueprints
app.register_blueprint(internal_bp)

# Context Processors
@app.context_processor
def utility_processor():
    def get_current_datetime_now():
        return datetime.datetime.now()
    return dict(now=get_current_datetime_now)

# Configure logging
if __name__ != '__main__':
    # When run by Gunicorn, use Gunicorn's logger
    gunicorn_logger = logging.getLogger('gunicorn.error')
    app.logger.handlers = gunicorn_logger.handlers
    app.logger.setLevel(gunicorn_logger.level)
else:
    # Basic configuration for development
    # logging.basicConfig(level=logging.DEBUG) # This is handled by Flask's default if run directly
    app.logger.setLevel(logging.DEBUG if app.debug else logging.INFO)


@app.errorhandler(401)
def unauthorized_error_handler(error):
    app.logger.warning(f"Unauthorized access attempt: {request.path} by user {request.headers.get('Remote-User', 'Unknown')}. IP: {request.remote_addr}")
    return render_template('errors/401.html'), 401

@app.errorhandler(403)
def forbidden_error_handler(error):
    app.logger.warning(
        f"Forbidden access attempt: {request.path} by user {request.headers.get('Remote-User', 'Unknown')}. "
        f"Description: {error.description if error else 'No description'}. IP: {request.remote_addr}"
    )
    return render_template('errors/403.html', description=error.description if error else "You do not have permission to access this page."), 403

@app.route('/')
def public_index():
    build_date_str = os.environ.get('APP_BUILD_DATE')
    if build_date_str:
        # Attempt to parse the build date if it's in a known format, e.g., ISO 8601
        # For simplicity, we'll use it as is, or format it if it's a parsable string.
        # If it's already formatted as desired by Docker build, just use it.
        # Example: If Docker passes "YYYY-MM-DD HH:MM:SS"
        last_updated = build_date_str
        # If Docker passes ISO 8601 (e.g., YYYY-MM-DDTHH:MM:SSZ) and you want to reformat:
        # try:
        #     dt_obj = datetime.datetime.fromisoformat(build_date_str.replace('Z', '+00:00'))
        #     last_updated = dt_obj.strftime("%Y-%m-%d %H:%M:%S UTC")
        # except ValueError:
        #     last_updated = build_date_str # Fallback to raw string if parsing fails
        app.logger.info(f"Using build date for last_updated: {last_updated}")
    else:
        # Fallback for local development or if APP_BUILD_DATE is not set
        now_dt = datetime.datetime.now(datetime.timezone.utc)
        last_updated = now_dt.strftime("%Y-%m-%d %H:%M:%S UTC")
        app.logger.info(f"APP_BUILD_DATE not set. Using current time for last_updated: {last_updated}")

    return render_template('index.html', last_updated=last_updated)

@app.route('/resume')
def resume_page():
    build_date_str = os.environ.get('APP_BUILD_DATE')
    if build_date_str:
        last_updated = build_date_str
    else:
        now_dt = datetime.datetime.now(datetime.timezone.utc)
        last_updated = now_dt.strftime("%Y-%m-%d %H:%M:%S UTC")
    return render_template('resume.html', last_updated=last_updated)

@app.route('/blog')
def blog_page():
    build_date_str = os.environ.get('APP_BUILD_DATE')
    if build_date_str:
        last_updated = build_date_str
    else:
        now_dt = datetime.datetime.now(datetime.timezone.utc)
        last_updated = now_dt.strftime("%Y-%m-%d %H:%M:%S UTC")
    return render_template('blog.html', last_updated=last_updated)

# The old /intern route is now handled by the internal_bp blueprint.

if __name__ == '__main__':
    # In production, Gunicorn will run the app. This is for development.
    # Ensure FLASK_ENV is set to 'development' for debug mode and dev headers to work
    if app.config['ENV'] == 'development':
        app.run(debug=True, host='0.0.0.0', port=os.environ.get('FLASK_RUN_PORT', 5000))
    else:
        # In a real production scenario, Gunicorn runs this, so app.run() isn't typically called.
        # However, if you were to run `python app.py` in a prod-like env without Gunicorn:
        app.run(host='0.0.0.0', port=os.environ.get('FLASK_RUN_PORT', 8000))
