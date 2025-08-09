# Simple test Dockerfile for demonstrating GeosChem build pipeline
FROM rockylinux:9

# Accept build arguments
ARG COMPILER=gcc
ARG COMPILER_VERSION=13
ARG ARCHITECTURE=x86_64

# Add metadata
LABEL maintainer="GeosChem AWS Platform"
LABEL architecture=${ARCHITECTURE}
LABEL compiler=${COMPILER}
LABEL compiler_version=${COMPILER_VERSION}

# Update system and install basic tools
RUN dnf update -y && \
    dnf install -y --allowerasing \
        gcc \
        gcc-gfortran \
        git \
        make \
        python3 \
        python3-pip \
        wget \
        curl && \
    dnf clean all

# Create a test application
WORKDIR /app
RUN echo '#!/bin/bash' > geoschem-test.sh && \
    echo 'echo "GeosChem Test Build Successful!"' >> geoschem-test.sh && \
    echo 'echo "Architecture: ${ARCHITECTURE}"' >> geoschem-test.sh && \
    echo 'echo "Compiler: ${COMPILER} ${COMPILER_VERSION}"' >> geoschem-test.sh && \
    echo 'echo "Build completed at $(date)"' >> geoschem-test.sh && \
    chmod +x geoschem-test.sh

# Set entrypoint
ENTRYPOINT ["/app/geoschem-test.sh"]
CMD []