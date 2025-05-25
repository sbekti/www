# Development Guide

This guide explains how to set up and run the Flask application in development mode.

## Quick Start

1. **Set up development environment:**
   ```bash
   ./setup_dev.sh
   ```

2. **Edit database configuration:**
   ```bash
   nano run_dev.sh  # or your preferred editor
   ```
   Update the database connection details with your PostgreSQL credentials.

3. **Run the development server:**
   ```bash
   ./run_dev.sh
   ```

## Development Scripts

### `setup_dev.sh`
Initializes the development environment by:
- Copying `run_dev.sh.template` to `run_dev.sh`
- Making it executable
- Providing setup instructions

### `run_dev.sh` (created from template)
Runs the Flask development server with:
- Proper environment variables
- Database connection configuration
- Automatic build date simulation
- Debug mode enabled
- Development headers simulation

**Note:** This file is ignored by git to protect your database credentials.

### `build_docker.sh`
Builds Docker images with automatic build date injection:
```bash
# Basic build
./build_docker.sh

# Custom image name and tag
./build_docker.sh my-flask-app v1.0

# Custom workers count
./build_docker.sh my-flask-app v1.0 8
```

## Docker Improvements

### Automatic Build Date
The Dockerfile now supports automatic build date population:

- **Build argument:** `BUILD_DATE` is automatically set by `build_docker.sh`
- **Environment variable:** `APP_BUILD_DATE` is available in the container
- **Fallback:** If not provided, the app uses current time

### Build Examples

```bash
# Using the build script (recommended)
./build_docker.sh www latest

# Manual build with custom date
docker build --build-arg BUILD_DATE="$(date -u +'%Y-%m-%d %H:%M:%S UTC')" -t www:latest .

# Build without date (app will use current time)
docker build -t www:latest .
```

## Environment Variables

### Development (`run_dev.sh`)
- `FLASK_APP=app.py`
- `FLASK_ENV=development`
- `DB_USERNAME` - PostgreSQL username
- `DB_PASSWORD` - PostgreSQL password
- `DB_HOST` - PostgreSQL host
- `DB_PORT` - PostgreSQL port
- `DB_NAME` - Database name
- `APP_BUILD_DATE` - Simulated build date
- `FLASK_RUN_PORT` - Optional custom port

### Production (Docker)
- `GUNICORN_WORKERS` - Number of Gunicorn workers (default: 4)
- `APP_BUILD_DATE` - Build timestamp
- Database variables (same as development)
- `SECRET_KEY` - Flask secret key (set in production!)

## Database Setup

Ensure your PostgreSQL database contains the required tables:
- `users`
- `radusergroup` 
- `radcheck`

## Security Notes

1. **`run_dev.sh` is git-ignored** to prevent credential leaks
2. **Use strong passwords** for database connections
3. **Set `SECRET_KEY`** environment variable in production
4. **Review database permissions** for production deployments

## Troubleshooting

### Database Connection Issues
1. Verify PostgreSQL is running
2. Check database credentials in `run_dev.sh`
3. Ensure database exists and is accessible
4. Check firewall/network settings

### Docker Build Issues
1. Ensure Docker is running
2. Check build script permissions: `chmod +x build_docker.sh`
3. Verify Dockerfile syntax

### Development Server Issues
1. Check Python dependencies: `pip install -r requirements.txt`
2. Verify script permissions: `chmod +x run_dev.sh`
3. Check port availability (default: 5000)

## File Structure

```
www/
├── app.py                 # Main Flask application
├── Dockerfile            # Improved with auto build date
├── build_docker.sh       # Docker build script
├── setup_dev.sh          # Development setup script
├── run_dev.sh.template   # Template for development script
├── run_dev.sh            # Your local dev script (git-ignored)
├── requirements.txt      # Python dependencies
├── .gitignore           # Updated to ignore run_dev.sh
└── DEVELOPMENT.md       # This file
``` 