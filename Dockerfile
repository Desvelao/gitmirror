FROM golang:1.18

# Create user with UID 1000
RUN useradd -u 1000 -m appuser

# Set HOME so Go uses /home/appuser/.cache/go-build
ENV HOME=/home/appuser

WORKDIR /app

# Switch to the new user
USER appuser
