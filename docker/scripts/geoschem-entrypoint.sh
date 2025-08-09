#!/bin/bash
# GeosChem Container Entrypoint - Supports both Classic and GCHP modes

set -e

# Default values
MODE=""
RESOLUTION=""
CORES=""
SIMULATION=""
CONFIG_DIR="/workspace/config"

print_usage() {
    echo "GeosChem Container - Supports both Classic and GCHP modes"
    echo ""
    echo "Usage:"
    echo "  # GeosChem Classic mode"
    echo "  $0 classic --simulation fullchem --resolution 4x5 [options]"
    echo ""  
    echo "  # GCHP High Performance mode"
    echo "  $0 gchp --simulation fullchem --resolution C48 --cores 24 [options]"
    echo ""
    echo "Arguments:"
    echo "  classic|gchp          Execution mode"
    echo "  --simulation TYPE     Simulation type (fullchem, aerosol, tropchem, etc.)"
    echo "  --resolution RES      Grid resolution:"
    echo "                        Classic: 4x5, 2x2.5, 0.5x0.625, 0.25x0.3125"  
    echo "                        GCHP: C48, C90, C180, C360, C720, C1440"
    echo "  --cores N             Number of MPI processes (GCHP only)"
    echo ""
    echo "Options:"
    echo "  --config-dir DIR      Configuration directory (default: /workspace/config)"
    echo "  --data-dir DIR        Input data directory (default: /workspace/data)"
    echo "  --output-dir DIR      Output directory (default: /workspace/output)"
    echo "  --start-date DATE     Start date (YYYY-MM-DD)"
    echo "  --end-date DATE       End date (YYYY-MM-DD)"
    echo "  --dry-run             Show commands without executing"
    echo "  --debug               Enable debug output"
    echo ""
    echo "Examples:"
    echo "  # Classic 4x5 full chemistry simulation"  
    echo "  $0 classic --simulation fullchem --resolution 4x5"
    echo ""
    echo "  # GCHP high-resolution run with 96 cores"
    echo "  $0 gchp --simulation fullchem --resolution C180 --cores 96"
    echo ""
    echo "  # Check available simulations"
    echo "  $0 list-simulations"
    echo ""
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        classic|gchp)
            MODE="$1"
            shift
            ;;
        --simulation)
            SIMULATION="$2"
            shift 2
            ;;
        --resolution)
            RESOLUTION="$2"  
            shift 2
            ;;
        --cores)
            CORES="$2"
            shift 2
            ;;
        --config-dir)
            CONFIG_DIR="$2"
            shift 2
            ;;
        --data-dir)
            DATA_DIR="$2"
            shift 2
            ;;
        --output-dir)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --start-date)
            START_DATE="$2"
            shift 2
            ;;
        --end-date)
            END_DATE="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=1
            shift
            ;;
        --debug)
            DEBUG=1
            set -x
            shift
            ;;
        --help)
            print_usage
            exit 0
            ;;
        list-simulations)
            echo "Available GeosChem simulations:"
            echo "  fullchem     - Full chemistry simulation"
            echo "  aerosol      - Aerosol-only simulation"
            echo "  tropchem     - Tropospheric chemistry"
            echo "  CH4          - Methane simulation"
            echo "  CO2          - Carbon dioxide simulation"
            echo "  TransportTracers - Transport tracer simulation"
            echo ""
            echo "Use --simulation <type> to specify"
            exit 0
            ;;
        *)
            echo "Unknown argument: $1"
            print_usage
            exit 1
            ;;
    esac
done

# Validate required arguments
if [[ -z "$MODE" ]]; then
    echo "Error: Must specify mode (classic or gchp)"
    print_usage
    exit 1
fi

if [[ -z "$SIMULATION" ]]; then
    echo "Error: Must specify simulation type (--simulation)"
    print_usage  
    exit 1
fi

if [[ -z "$RESOLUTION" ]]; then
    echo "Error: Must specify resolution (--resolution)"
    print_usage
    exit 1
fi

# Mode-specific validation
if [[ "$MODE" == "gchp" && -z "$CORES" ]]; then
    echo "Error: GCHP mode requires --cores specification"
    print_usage
    exit 1
fi

# Set default directories if not specified
DATA_DIR="${DATA_DIR:-/workspace/data}"
OUTPUT_DIR="${OUTPUT_DIR:-/workspace/output}"

# Create output directories
mkdir -p "$OUTPUT_DIR" "$CONFIG_DIR"

# Load environment
source /opt/spack/share/spack/setup-env.sh

echo "================================================"
echo "GeosChem Container Runtime"
echo "================================================"
echo "Mode: $MODE"
echo "Simulation: $SIMULATION"
echo "Resolution: $RESOLUTION"
[[ "$MODE" == "gchp" ]] && echo "Cores: $CORES"
echo "Config: $CONFIG_DIR"
echo "Data: $DATA_DIR"
echo "Output: $OUTPUT_DIR"
echo "================================================"

# Execute the appropriate runner
if [[ "$MODE" == "classic" ]]; then
    exec /usr/local/bin/run-classic.sh \
        --simulation "$SIMULATION" \
        --resolution "$RESOLUTION" \
        --config-dir "$CONFIG_DIR" \
        --data-dir "$DATA_DIR" \
        --output-dir "$OUTPUT_DIR" \
        ${START_DATE:+--start-date "$START_DATE"} \
        ${END_DATE:+--end-date "$END_DATE"} \
        ${DRY_RUN:+--dry-run}
        
elif [[ "$MODE" == "gchp" ]]; then
    exec /usr/local/bin/run-gchp.sh \
        --simulation "$SIMULATION" \
        --resolution "$RESOLUTION" \
        --cores "$CORES" \
        --config-dir "$CONFIG_DIR" \
        --data-dir "$DATA_DIR" \
        --output-dir "$OUTPUT_DIR" \
        ${START_DATE:+--start-date "$START_DATE"} \
        ${END_DATE:+--end-date "$END_DATE"} \
        ${DRY_RUN:+--dry-run}
        
else
    echo "Error: Unknown mode '$MODE'"
    exit 1
fi