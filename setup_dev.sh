#!/bin/bash

# Development environment setup script
# This script helps set up the development environment for the Flask app

set -e

echo "🚀 Setting up Flask development environment..."
echo ""

# Check if run_dev.sh already exists
if [ -f "run_dev.sh" ]; then
    echo "⚠️  run_dev.sh already exists!"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Setup cancelled. Your existing run_dev.sh was not modified."
        exit 0
    fi
fi

# Copy template to run_dev.sh
echo "📋 Copying run_dev.sh.template to run_dev.sh..."
cp run_dev.sh.template run_dev.sh
chmod +x run_dev.sh

echo "✅ run_dev.sh created successfully!"
echo ""

echo "📝 Next steps:"
echo "1. Edit run_dev.sh and update the database connection details:"
echo "   - DB_USERNAME: Your PostgreSQL username"
echo "   - DB_PASSWORD: Your PostgreSQL password"
echo "   - DB_HOST: Your PostgreSQL host (usually 'localhost')"
echo "   - DB_PORT: Your PostgreSQL port (usually '5432')"
echo "   - DB_NAME: Your database name"
echo ""
echo "2. Make sure your PostgreSQL database is running and accessible"
echo ""
echo "3. Install Python dependencies (if not already done):"
echo "   pip install -r requirements.txt"
echo ""
echo "4. Run the development server:"
echo "   ./run_dev.sh"
echo ""

echo "🔒 Security note:"
echo "   run_dev.sh is already added to .gitignore to prevent"
echo "   your database credentials from being committed to git."
echo ""

echo "🐳 Docker build:"
echo "   To build a Docker image with automatic build date:"
echo "   ./build_docker.sh [image_name] [tag] [workers]"
echo ""

echo "✨ Setup complete! Happy coding!" 