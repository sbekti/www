# Use a Python base image
FROM python:3.9-alpine

# Build arguments with defaults
ARG GUNICORN_WORKERS=4
ARG BUILD_DATE

# Set environment variables
ENV GUNICORN_WORKERS=$GUNICORN_WORKERS

# Set build date environment variable
# If BUILD_DATE is not provided during build, it will be empty and app will use current time
ENV APP_BUILD_DATE=$BUILD_DATE

# Set the working directory in the container
WORKDIR /app

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S -G appgroup appuser

# Copy the requirements file first to leverage Docker cache
COPY requirements.txt /app/

# Install the required Python packages
RUN pip install --no-cache-dir -r requirements.txt

# Copy the rest of the application files
COPY . /app

# Change ownership of the app directory to the new user
RUN chown -R appuser:appgroup /app

# Switch to the non-root user
USER appuser

# Expose the port that Gunicorn will listen on
EXPOSE 8000

# Health check to ensure the application is running
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8000/ || exit 1

# Command to run the Flask application using Gunicorn
CMD ["sh", "-c", "gunicorn -w $GUNICORN_WORKERS -b 0.0.0.0:8000 app:app"]
