#!/bin/bash
# Configure GeosChem data sources using AWS Open Data Archive

set -e

# GeosChem AWS Open Data Registry URLs
# Reference: https://registry.opendata.aws/geoschem/
GEOSCHEM_S3_BASE="s3://gcgrid"
GEOSCHEM_HTTP_BASE="https://gcgrid.s3.amazonaws.com"

# Input data categories available on AWS Open Data
declare -A DATA_SOURCES=(
    ["MetDir"]="GEOS_4x5 GEOS_2x2.5 GEOS_0.5x0.625 GEOS_0.25x0.3125"
    ["HEMCO"]="HEMCO"  
    ["CHEM_INPUTS"]="CHEM_INPUTS"
    ["ExtData"]="ExtData"
    ["BenchmarkDataDir"]="BenchmarkDataDir"
)

print_usage() {
    echo "Configure GeosChem data sources from AWS Open Data Archive"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --resolution RES      Target resolution (4x5, 2x2.5, 0.5x0.625, 0.25x0.3125, C48, C90, C180, C360)"
    echo "  --data-dir DIR        Local data directory (default: /workspace/data)"
    echo "  --mode classic|gchp   Execution mode"
    echo "  --sync                Sync data locally (default: direct S3 access)"
    echo "  --list                List available datasets"
    echo "  --dry-run             Show what would be configured"
    echo ""
    echo "Examples:"
    echo "  # Configure for classic 4x5 simulation"
    echo "  $0 --resolution 4x5 --mode classic"
    echo ""
    echo "  # Configure for GCHP C180 simulation"  
    echo "  $0 --resolution C180 --mode gchp"
    echo ""
    echo "  # Sync specific data locally"
    echo "  $0 --resolution 2x2.5 --mode classic --sync"
    echo ""
}

list_datasets() {
    echo "Available GeosChem datasets on AWS Open Data Archive:"
    echo ""
    echo "Meteorological Data (MetDir):"
    echo "  GEOS_4x5/         - 4°x5° resolution"
    echo "  GEOS_2x2.5/       - 2°x2.5° resolution"
    echo "  GEOS_0.5x0.625/   - 0.5°x0.625° resolution"
    echo "  GEOS_0.25x0.3125/ - 0.25°x0.3125° resolution"
    echo ""
    echo "Emission Data (HEMCO):"
    echo "  HEMCO/            - Emission inventories and scaling factors"
    echo ""
    echo "Chemistry Data:"
    echo "  CHEM_INPUTS/      - Chemical mechanism data, photolysis, etc."
    echo ""
    echo "External Data (for GCHP):"
    echo "  ExtData/          - GCHP-specific input data"
    echo ""
    echo "Benchmark Data:"
    echo "  BenchmarkDataDir/ - Reference benchmark datasets"
    echo ""
    echo "Total archive size: ~20TB"
    echo "HTTP access: $GEOSCHEM_HTTP_BASE"
    echo "S3 access: $GEOSCHEM_S3_BASE"
}

configure_classic_data() {
    local resolution=$1
    local data_dir=$2
    local sync_mode=$3
    
    echo "Configuring GeosChem Classic data for resolution: $resolution"
    
    # Map resolution to met directory
    local met_dir=""
    case $resolution in
        4x5) met_dir="GEOS_4x5" ;;
        2x2.5) met_dir="GEOS_2x2.5" ;;
        0.5x0.625) met_dir="GEOS_0.5x0.625" ;;
        0.25x0.3125) met_dir="GEOS_0.25x0.3125" ;;
        *) echo "Error: Unsupported classic resolution: $resolution"; return 1 ;;
    esac
    
    mkdir -p "$data_dir"
    
    if [[ "$sync_mode" == "true" ]]; then
        echo "Syncing data locally (this may take significant time and storage)..."
        # Sync specific directories only
        aws s3 sync "${GEOSCHEM_S3_BASE}/${met_dir}/" "${data_dir}/${met_dir}/" --no-progress
        aws s3 sync "${GEOSCHEM_S3_BASE}/HEMCO/" "${data_dir}/HEMCO/" --no-progress
        aws s3 sync "${GEOSCHEM_S3_BASE}/CHEM_INPUTS/" "${data_dir}/CHEM_INPUTS/" --no-progress
    else
        echo "Configuring direct S3 access (recommended)..."
        # Create symbolic links or configuration for direct S3 access
        echo "export GEOSCHEM_DATA_ROOT='${GEOSCHEM_S3_BASE}'" >> /workspace/geoschem-env.sh
        echo "export GEOSCHEM_MET_DIR='${GEOSCHEM_S3_BASE}/${met_dir}'" >> /workspace/geoschem-env.sh
        echo "export GEOSCHEM_HEMCO_DIR='${GEOSCHEM_S3_BASE}/HEMCO'" >> /workspace/geoschem-env.sh
        echo "export GEOSCHEM_CHEM_INPUTS='${GEOSCHEM_S3_BASE}/CHEM_INPUTS'" >> /workspace/geoschem-env.sh
    fi
    
    echo "Classic data configuration complete for $resolution"
}

configure_gchp_data() {
    local resolution=$1
    local data_dir=$2
    local sync_mode=$3
    
    echo "Configuring GCHP data for resolution: $resolution"
    
    # Validate GCHP resolution
    case $resolution in
        C48|C90|C180|C360|C720|C1440) ;;
        *) echo "Error: Unsupported GCHP resolution: $resolution"; return 1 ;;
    esac
    
    mkdir -p "$data_dir"
    
    if [[ "$sync_mode" == "true" ]]; then
        echo "Syncing GCHP data locally..."
        # GCHP typically needs higher resolution met data
        aws s3 sync "${GEOSCHEM_S3_BASE}/GEOS_0.25x0.3125/" "${data_dir}/GEOS_0.25x0.3125/" --no-progress
        aws s3 sync "${GEOSCHEM_S3_BASE}/ExtData/" "${data_dir}/ExtData/" --no-progress
        aws s3 sync "${GEOSCHEM_S3_BASE}/HEMCO/" "${data_dir}/HEMCO/" --no-progress
        aws s3 sync "${GEOSCHEM_S3_BASE}/CHEM_INPUTS/" "${data_dir}/CHEM_INPUTS/" --no-progress
    else
        echo "Configuring direct S3 access for GCHP..."
        echo "export GCHP_DATA_ROOT='${GEOSCHEM_S3_BASE}'" >> /workspace/geoschem-env.sh
        echo "export GCHP_EXTDATA_DIR='${GEOSCHEM_S3_BASE}/ExtData'" >> /workspace/geoschem-env.sh
        echo "export GCHP_MET_DIR='${GEOSCHEM_S3_BASE}/GEOS_0.25x0.3125'" >> /workspace/geoschem-env.sh
        echo "export GCHP_HEMCO_DIR='${GEOSCHEM_S3_BASE}/HEMCO'" >> /workspace/geoschem-env.sh
        echo "export GCHP_CHEM_INPUTS='${GEOSCHEM_S3_BASE}/CHEM_INPUTS'" >> /workspace/geoschem-env.sh
    fi
    
    echo "GCHP data configuration complete for $resolution"
}

estimate_data_size() {
    local resolution=$1
    local mode=$2
    
    echo "Estimated data requirements for $mode mode at $resolution resolution:"
    
    case "$mode-$resolution" in
        classic-4x5) echo "  ~50GB for 1-year simulation" ;;
        classic-2x2.5) echo "  ~200GB for 1-year simulation" ;;
        classic-0.5x0.625) echo "  ~2TB for 1-year simulation" ;;
        classic-0.25x0.3125) echo "  ~8TB for 1-year simulation" ;;
        gchp-C48) echo "  ~100GB for 1-year simulation" ;;
        gchp-C90) echo "  ~400GB for 1-year simulation" ;;
        gchp-C180) echo "  ~1.5TB for 1-year simulation" ;;
        gchp-C360) echo "  ~6TB for 1-year simulation" ;;
        *) echo "  Unable to estimate for $mode-$resolution" ;;
    esac
    
    echo ""
    echo "Note: Direct S3 access eliminates local storage requirements"
    echo "      and provides faster access in most AWS regions."
}

# Parse command line arguments
RESOLUTION=""
DATA_DIR="/workspace/data"
MODE=""
SYNC_MODE="false"
DRY_RUN="false"

while [[ $# -gt 0 ]]; do
    case $1 in
        --resolution) RESOLUTION="$2"; shift 2 ;;
        --data-dir) DATA_DIR="$2"; shift 2 ;;
        --mode) MODE="$2"; shift 2 ;;
        --sync) SYNC_MODE="true"; shift ;;
        --list) list_datasets; exit 0 ;;
        --dry-run) DRY_RUN="true"; shift ;;
        --help) print_usage; exit 0 ;;
        *) echo "Unknown argument: $1"; print_usage; exit 1 ;;
    esac
done

# Validate required arguments
if [[ -z "$RESOLUTION" || -z "$MODE" ]]; then
    echo "Error: --resolution and --mode are required"
    print_usage
    exit 1
fi

# Show data size estimates
estimate_data_size "$RESOLUTION" "$MODE"

if [[ "$DRY_RUN" == "true" ]]; then
    echo "DRY RUN - would configure:"
    echo "  Resolution: $RESOLUTION"
    echo "  Mode: $MODE"
    echo "  Data directory: $DATA_DIR"
    echo "  Sync mode: $SYNC_MODE"
    exit 0
fi

# Configure data sources based on mode
case $MODE in
    classic)
        configure_classic_data "$RESOLUTION" "$DATA_DIR" "$SYNC_MODE"
        ;;
    gchp)
        configure_gchp_data "$RESOLUTION" "$DATA_DIR" "$SYNC_MODE"
        ;;
    *)
        echo "Error: Mode must be 'classic' or 'gchp'"
        exit 1
        ;;
esac

echo ""
echo "GeosChem data configuration complete!"
echo "Environment variables saved to /workspace/geoschem-env.sh"
echo ""
echo "To use: source /workspace/geoschem-env.sh"