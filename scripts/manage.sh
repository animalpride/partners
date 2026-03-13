#!/bin/bash
set -e

# ==============================================================================
# DenOps Monorepo Management Script
#
# This script is the single entry point for managing both the local development
# environment and deploying services to the Swarm cluster.
#
# Commands:
#   up      - Start the local development environment in detached mode.
#   down    - Stop and remove the local development environment.
#   logs    - Tail the logs of the local development environment.
#   deploy  - Build, push, and deploy changed services to the Swarm cluster.
#
# Usage:
#   ./scripts/manage.sh [command]
#
# ==============================================================================

# --- Configuration ---
DEV_COMPOSE_FILE="development/docker-compose.yml"
BASE_COMPOSE_FILE="deployment/base/docker-compose.yml"
QA_COMPOSE_FILE="deployment/qa/docker-compose.override.yml"
PROD_COMPOSE_FILE="deployment/prod/docker-compose.override.yml"

# Load environment variables from .env file if it exists in the scripts directory
if [ -f "$(dirname "$0")/.env" ]; then
  set -a # Automatically export all variables
  source "$(dirname "$0")/.env"
  set +a
fi

# Service definitions: "service_name:directory:image_name"
# The image name now uses the REGISTRY_URL variable.
SERVICES=(
  "auth:services/auth:${REGISTRY_URL:-registry.octagon.home}/animal-pride/partners-auth"
  "ui:services/ui:${REGISTRY_URL:-registry.octagon.home}/animal-pride/partners-ui"
  "core:services/core:${REGISTRY_URL:-registry.octagon.home}/animal-pride/partners-core"
)
MIGRATE_SERVICES=("auth" "core")

QA_SWARM_CONTEXT="octagon-swarm"
PROD_SWARM_CONTEXT="partners-animalpride-prod"
DEFAULT_CONTEXT="default"
STACK_NAME="partners"
CONFIG_KEEP_DEFAULT=2
MIGRATE_ENV_DEV_FILE="development/.env"
MIGRATE_ENV_QA_FILE="deployment/qa/.env"
MIGRATE_ENV_PROD_FILE="deployment/prod/.env"
MIGRATE_ENV_ACTIVE_FILE=""

# --- Helper Functions ---
usage() {
  echo "Usage: $0 {up|down|logs|build-push|deploy|migrate|logout}"
  echo "Commands:"
  echo "  up             - Start the local development environment."
  echo "  down           - Stop the local development environment."
  echo "  logs [svc]     - Tail logs for all or a specific service."
  echo "  build-push [tag] [svc...] - Build and push images. If tag provided, push tag and latest."
  echo "  deploy <qa|prod> <tag>   - Deploy to Swarm using the target compose file and image tag."
  echo "  migrate <qa|prod> <up|down> [service] [version] - Migrate databases for the environment."
  echo "  migrate <qa|prod> force <service> <version>     - Force-set migration version (clears dirty)."
  echo ""
  echo "Migration usage:"
  echo "  $0 migrate <qa|prod> up [service] [version]      # Migrate up to latest or specified version."
  echo "  $0 migrate <qa|prod> down <service> <N>          # Migrate down N steps (N=number of migrations to revert, or 0 for all)."
  echo "  $0 migrate <qa|prod> force <service> <version>   # Set version and clear dirty flag."
  echo "  Valid services: auth, core, events, assets, inventory, sales-channel"
  echo ""
  echo "Examples:"
  echo "  $0 build-push          # Build and push all changed services (latest)"
  echo "  $0 build-push v1.2.3   # Build and push tag v1.2.3 and latest"
  echo "  $0 build-push v1.2.3 ui # Build and push tag v1.2.3 and latest for ui"
  echo "  $0 deploy qa v1.2.3    # Deploy QA stack using image tag v1.2.3"
  echo "  $0 deploy prod v1.2.3  # Deploy prod stack using image tag v1.2.3"
  echo "  $0 migrate qa up               # Migrate all services up to latest"
  echo "  $0 migrate qa up auth           # Migrate auth service up to latest"
  echo "  $0 migrate qa up auth 000001    # Migrate auth service up to version 000001"
  echo "  $0 migrate prod down events 1   # Revert the last migration for events service"
  echo "  $0 migrate qa force core 2      # Force core DB to version 2 (clears dirty)"
  echo ""
  echo "Environment Variables:"
  echo "  IMAGE_TAG=v1.2.3             # Optional CI tag fallback for build-push"
  echo "  SKIP_IMAGE_CHECK=true           # Skip image existence verification (useful for prod/VPN issues)"
  echo "  IMAGE_CHECK_TIMEOUT=30          # Timeout in seconds for image checks (default: 10)"
  echo ""
  echo "Examples with environment variables:"
  echo "  SKIP_IMAGE_CHECK=true $0 deploy prod v1.2.3        # Deploy without checking if images exist"
  echo "  IMAGE_CHECK_TIMEOUT=30 $0 deploy prod v1.2.3       # Use 30s timeout for image checks"
  exit 1
}

# Check for required environment variables for deployment
check_registry_env_vars() {
  local missing_vars=()
  check_registry_url
  if [ -z "$REGISTRY_USER" ]; then missing_vars+=("REGISTRY_USER"); fi
  if [ -z "$REGISTRY_PASSWORD" ]; then missing_vars+=("REGISTRY_PASSWORD"); fi

  if [ ${#missing_vars[@]} -ne 0 ]; then
    echo "Error: The following required registry environment variables are not set:"
    for var in "${missing_vars[@]}"; do
      echo "  - $var"
    done
    echo "Please create a 'scripts/.env' file based on 'scripts/.env.example' and define them."
    exit 1
  fi
}

check_registry_url() {
  if [ -z "$REGISTRY_URL" ]; then
    echo "Error: REGISTRY_URL is not set."
    echo "Please set REGISTRY_URL in scripts/.env before deploying."
    exit 1
  fi
}

require_image_tag() {
  local tag="$1"

  if [ -z "$tag" ]; then
    echo "Error: Missing image tag."
    echo "Usage: $0 deploy <qa|prod> <tag>"
    exit 1
  fi
}

resolve_ci_image_tag() {
  if [ -n "${IMAGE_TAG:-}" ]; then
    echo "$IMAGE_TAG"
    return 0
  fi

  if [ -n "${CI_COMMIT_TAG:-}" ]; then
    echo "$CI_COMMIT_TAG"
    return 0
  fi

  if [ "${GITHUB_REF_TYPE:-}" = "tag" ] && [ -n "${GITHUB_REF_NAME:-}" ]; then
    echo "$GITHUB_REF_NAME"
    return 0
  fi

  if [ -n "${BITBUCKET_TAG:-}" ]; then
    echo "$BITBUCKET_TAG"
    return 0
  fi

  if [ -n "${DRONE_TAG:-}" ]; then
    echo "$DRONE_TAG"
    return 0
  fi

  if [ -n "${CI_COMMIT_SHORT_SHA:-}" ]; then
    echo "sha-${CI_COMMIT_SHORT_SHA}"
    return 0
  fi

  if [ -n "${GITHUB_SHA:-}" ]; then
    echo "sha-$(echo "$GITHUB_SHA" | cut -c1-12)"
    return 0
  fi

  if [ -n "${BITBUCKET_COMMIT:-}" ]; then
    echo "sha-$(echo "$BITBUCKET_COMMIT" | cut -c1-12)"
    return 0
  fi

  if [ -n "${DRONE_COMMIT_SHA:-}" ]; then
    echo "sha-$(echo "$DRONE_COMMIT_SHA" | cut -c1-12)"
    return 0
  fi

  if [ -n "${BUILD_SOURCEVERSION:-}" ]; then
    echo "sha-$(echo "$BUILD_SOURCEVERSION" | cut -c1-12)"
    return 0
  fi

  return 1
}

require_images_exist() {
  local context="$1"
  local tag="$2"
  local missing_images=()
  local timeout="${IMAGE_CHECK_TIMEOUT:-10}"

  # Allow skipping image check via environment variable (useful for prod deployments with network issues)
  if [ "${SKIP_IMAGE_CHECK:-false}" == "true" ]; then
    echo "Skipping image existence check (SKIP_IMAGE_CHECK=true)"
    return 0
  fi

  echo "Verifying images exist in registry (timeout: ${timeout}s per image)..."
  for service_config in "${SERVICES[@]}"; do
    IFS=':' read -r _ _ image_name <<< "$service_config"
    echo -n "  Checking ${image_name}:${tag}... "
    
    # Use timeout to prevent hanging indefinitely
    # manifest inspect is a client-side registry operation; context is not applicable here
    if timeout "$timeout" docker manifest inspect "${image_name}:${tag}" > /dev/null 2>&1; then
      echo "✓"
    else
      if [ $? -eq 124 ]; then
        echo "⏱ timeout (${timeout}s)"
        echo "Warning: Image check timed out. Set SKIP_IMAGE_CHECK=true to bypass, or increase IMAGE_CHECK_TIMEOUT."
      else
        echo "✗"
      fi
      missing_images+=("${image_name}:${tag}")
    fi
  done

  if [ ${#missing_images[@]} -ne 0 ]; then
    echo ""
    echo "Error: The following images/tags do not exist or could not be verified:"
    for image in "${missing_images[@]}"; do
      echo "  - ${image}"
    done
    echo ""
    echo "Options:"
    echo "  1. Build and push the images first: $0 build-push $tag"
    echo "  2. Skip this check: SKIP_IMAGE_CHECK=true $0 deploy <env> $tag"
    echo "  3. Increase timeout: IMAGE_CHECK_TIMEOUT=30 $0 deploy <env> $tag"
    exit 1
  fi
  
  echo "All images verified successfully."
}

check_migrate_env_vars() {
  local missing_vars=()
  if [ -z "$DB_HOST" ]; then missing_vars+=("DB_HOST"); fi
  if [ -z "$DB_PORT" ]; then missing_vars+=("DB_PORT"); fi
  if [ -z "$DB_USER" ]; then missing_vars+=("DB_USER"); fi
  if [ -z "$DB_PASSWORD" ]; then missing_vars+=("DB_PASSWORD"); fi
  if [ -z "$AUTH_SERVICE_DB_NAME" ]; then missing_vars+=("AUTH_SERVICE_DB_NAME"); fi
  if [ -z "$CORE_SERVICE_DB_NAME" ]; then missing_vars+=("CORE_SERVICE_DB_NAME"); fi


  if [ ${#missing_vars[@]} -ne 0 ]; then
    echo "Error: The following required database environment variables are not set:"
    for var in "${missing_vars[@]}"; do
      echo "  - $var"
    done
    if [ -n "$MIGRATE_ENV_ACTIVE_FILE" ]; then
      echo "Update $MIGRATE_ENV_ACTIVE_FILE with these values."
    else
      echo "Please create a deployment environment file with these values."
    fi
    exit 1
  fi
}

load_migrate_env() {
  local target_env="$1"
  local env_file=""

  if [ "$target_env" == "dev" ]; then
    env_file="$MIGRATE_ENV_DEV_FILE"
  elif [ "$target_env" == "qa" ]; then
    env_file="$MIGRATE_ENV_QA_FILE"
  elif [ "$target_env" == "prod" ]; then
    env_file="$MIGRATE_ENV_PROD_FILE"
  else
    echo "Error: Missing or invalid migrate target. Use 'dev', 'qa' or 'prod'."
    usage
  fi

  if [ ! -f "$env_file" ]; then
    echo "Error: Migrate env file not found at $env_file"
    echo "Create it from ${env_file}.example"
    exit 1
  fi

  MIGRATE_ENV_ACTIVE_FILE="$env_file"

  set -a
  source "$env_file"
  set +a

  check_migrate_env_vars
}

resolve_migration_path() {
  local service="$1"
  local migration_path=""

  for service_config in "${SERVICES[@]}"; do
    IFS=':' read -r svc_name svc_dir image_name <<< "$service_config"
    if [ "$svc_name" == "$service" ]; then
      migration_path="$svc_dir/migrations"
      break
    fi
  done

  echo "$migration_path"
}

run_service_migration() {
  local service="$1"
  local direction="$2"
  local version="$3"
  local db_name="$4"
  local migration_path=""
  local DATABASE_URL=""
  local MIGRATION_DIR=""

  migration_path=$(resolve_migration_path "$service")
  if [ -z "$migration_path" ]; then
    echo "Error: Migration path not found for service '$service'."
    exit 1
  fi

  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  REPO_ROOT="$SCRIPT_DIR/.."
  MIGRATION_DIR="$REPO_ROOT/$migration_path"

  if [ ! -d "$MIGRATION_DIR" ]; then
    echo "Error: Migration directory '$migration_path' does not exist."
    exit 1
  fi

  DATABASE_URL="mysql://${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${db_name}?parseTime=true"

  if [ "$direction" == "up" ]; then
    if [ -z "$version" ]; then
      echo "No version specified. Will migrate up all pending migrations."
      echo "Migrating database for service '$service' (DB: $db_name) direction: $direction (all)"
      docker run --rm \
        -v "$MIGRATION_DIR:/migrations" \
        migrate/migrate \
        -path=/migrations -database="$DATABASE_URL" "$direction"
      return
    fi
  fi

  if [ -z "$version" ]; then
    echo "Error: Version must be specified for direction '$direction'."
    exit 1
  fi

  echo "Migrating database for service '$service' (DB: $db_name) direction: $direction version: $version"
  docker run --rm \
    -v "$MIGRATION_DIR:/migrations" \
    migrate/migrate \
    -path=/migrations -database="$DATABASE_URL" "$direction" "$version"
}

# Log in to the container registry
registry_login() {
  echo "--- Logging into container registry at $REGISTRY_URL ---"
  if echo "$REGISTRY_PASSWORD" | docker login "$REGISTRY_URL" -u "$REGISTRY_USER" --password-stdin; then
    echo "Registry login successful."
  else
    echo "Error: Registry login failed. Please check your credentials in 'scripts/.env'."
    exit 1
  fi
}

registry_login_context() {
  local context="$1"
  echo "--- Logging into container registry at $REGISTRY_URL (context: $context) ---"
  if echo "$REGISTRY_PASSWORD" | docker --context "$context" login "$REGISTRY_URL" -u "$REGISTRY_USER" --password-stdin; then
    echo "Registry login successful."
  else
    echo "Error: Registry login failed. Please check your credentials in 'scripts/.env'."
    exit 1
  fi
}

# Log out from the container registry
registry_logout() {
  if [ -n "$REGISTRY_URL" ]; then
    echo "--- Logging out from container registry at $REGISTRY_URL ---"
    docker logout "$REGISTRY_URL"
    echo "Logout complete."
  fi
}

registry_logout_context() {
  local context="$1"
  if [ -n "$REGISTRY_URL" ]; then
    echo "--- Logging out from container registry at $REGISTRY_URL (context: $context) ---"
    docker --context "$context" logout "$REGISTRY_URL"
    echo "Logout complete."
  fi
}

registry_auth_present() {
  local registry="$1"
  local docker_config_dir="${DOCKER_CONFIG:-$HOME/.docker}"
  local config_file="$docker_config_dir/config.json"
  local python_bin=""

  if [ ! -f "$config_file" ]; then
    return 1
  fi

  if command -v python3 >/dev/null 2>&1; then
    python_bin="python3"
  elif command -v python >/dev/null 2>&1; then
    python_bin="python"
  fi

  if [ -n "$python_bin" ]; then
    "$python_bin" - "$registry" "$config_file" <<'PY'
import json
import subprocess
import sys

registry = sys.argv[1]
config_file = sys.argv[2]

try:
    with open(config_file, "r", encoding="utf-8") as handle:
        cfg = json.load(handle)
except Exception:
    sys.exit(2)

auths = cfg.get("auths", {}) or {}
cred_helpers = cfg.get("credHelpers", {}) or {}
creds_store = cfg.get("credsStore") or ""

def has_auth_entry() -> bool:
    if registry in auths:
        return True
    for key in auths.keys():
        if registry in key:
            return True
    return False

def helper_has_registry(helper: str) -> bool:
    cmd = f"docker-credential-{helper}"
    try:
        result = subprocess.run(
            [cmd, "list"],
            check=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )
    except Exception:
        return False
    return registry in (result.stdout or "")

if has_auth_entry():
    sys.exit(0)

if registry in cred_helpers:
    if helper_has_registry(cred_helpers[registry]):
        sys.exit(0)

if creds_store:
    if helper_has_registry(creds_store):
        sys.exit(0)

sys.exit(1)
PY
    return $?
  fi

  grep -q "\"$registry\"" "$config_file"
}

require_registry_login_context() {
  local context="$1"

  check_registry_url

  if registry_auth_present "$REGISTRY_URL"; then
    return 0
  fi

  echo "Error: Registry credentials for $REGISTRY_URL are not available in the Docker client config."
  echo "Run: docker --context \"$context\" login \"$REGISTRY_URL\""
  exit 1
}

compute_config_hash() {
  local file_path="$1"
  local hash=""

  hash=$(sha256sum "$file_path" | awk '{print $1}' | cut -c1-8)
  if [ -z "$hash" ]; then
    echo "Error: Unable to compute hash for $file_path"
    exit 1
  fi

  echo "$hash"
}

create_swarm_config() {
  local context="$1"
  local name="$2"
  local file_path="$3"
  local env="$4"
  local role="$5"
  local version="$6"

  if docker --context "$context" config inspect "$name" &> /dev/null; then
    echo "Config $name already exists; reusing."
    return
  fi

  docker --context "$context" config create \
    --label "partners.config=true" \
    --label "partners.env=$env" \
    --label "partners.role=$role" \
    --label "partners.version=$version" \
    "$name" "$file_path"
}

cleanup_swarm_configs() {
  local context="$1"
  local env="$2"
  local role="$3"
  local keep_count="$4"
  local ids=()
  local entries=()

  mapfile -t ids < <(
    docker --context "$context" config ls \
      --filter "label=partners.config=true" \
      --filter "label=partners.env=$env" \
      --filter "label=partners.role=$role" \
      --format '{{.ID}}'
  )

  if [ ${#ids[@]} -le "$keep_count" ]; then
    return
  fi

  mapfile -t entries < <(
    for id in "${ids[@]}"; do
      docker --context "$context" config inspect -f '{{.CreatedAt}}|{{.Spec.Name}}' "$id"
    done | sort -r
  )

  local index=0
  for entry in "${entries[@]}"; do
    index=$((index + 1))
    if [ "$index" -le "$keep_count" ]; then
      continue
    fi

    local name
    name=$(echo "$entry" | cut -d'|' -f2)
    if ! docker --context "$context" config rm "$name"; then
      echo "Warning: Unable to remove config $name (in use or missing)."
    fi
  done
}

# --- Command Functions ---

# Start the development environment
dev_up() {
  echo "Starting local development environment..."
  docker compose -f "$DEV_COMPOSE_FILE" up -d
  echo "Development environment is up and running."
}

# Stop the development environment
dev_down() {
  echo "Stopping local development environment..."
  docker compose -f "$DEV_COMPOSE_FILE" down
  echo "Development environment has been stopped."
}

# Tail logs from the development environment
dev_logs() {
  echo "Tailing logs... (Press Ctrl+C to exit)"
  if [ -z "$1" ]; then
    docker compose -f "$DEV_COMPOSE_FILE" logs -f
  else
    docker compose -f "$DEV_COMPOSE_FILE" logs -f "$1"
  fi
}

# Build and push images to registry (QA only)
build_push_services() {
  local maybe_tag="$1"
  local force_services=()
  local tag=""
  local services_found=()
  local services_not_found=()

  if [ -n "$maybe_tag" ]; then
    local is_service=false
    for service_config in "${SERVICES[@]}"; do
      IFS=':' read -r service_name service_dir image_name <<< "$service_config"
      if [ "$service_name" == "$maybe_tag" ]; then
        is_service=true
        break
      fi
    done

    if [ "$is_service" = false ]; then
      tag="$maybe_tag"
      shift
    fi
  fi

  if [ -z "$tag" ]; then
    if resolved_tag=$(resolve_ci_image_tag); then
      tag="$resolved_tag"
      echo "No explicit tag provided. Resolved CI image tag: $tag"
    fi
  fi

  force_services=("$@")

  check_registry_env_vars
  require_registry_login_context "default"

  # --- Build and Push Logic ---
  if [ ${#force_services[@]} -gt 0 ]; then
    # Specific services are requested, so we force the build for each.
    echo "--- Force build/push for services: ${force_services[*]} ---"
    for requested_service in "${force_services[@]}"; do
      local service_found=false
      for service_config in "${SERVICES[@]}"; do
        IFS=':' read -r service_name service_dir image_name <<< "$service_config"
        if [ "$service_name" == "$requested_service" ]; then
          service_found=true
          services_found+=("$service_name")
          # Go services need the services/ parent as build context so the shared
          # module (referenced via replace ../shared in go.mod) is accessible.
          local build_context="$service_dir"
          local -a build_file_args=()
          local -a build_arg_flags=()
          if [ "$service_name" != "ui" ]; then
            build_context="services"
            build_file_args=("-f" "$service_dir/Dockerfile")
          elif [ -n "$tag" ]; then
            build_arg_flags=("--build-arg" "VITE_APP_VERSION=$tag")
          fi
          if [ -n "$tag" ]; then
            echo "Building and pushing image: $image_name:$tag and $image_name:latest"
            docker build "${build_file_args[@]}" "${build_arg_flags[@]}" -t "$image_name:latest" -t "$image_name:$tag" "$build_context"
            docker push "$image_name:$tag"
            docker push "$image_name:latest"
          else
            echo "Building and pushing image: $image_name:latest"
            docker build "${build_file_args[@]}" "${build_arg_flags[@]}" -t "$image_name:latest" "$build_context"
            docker push "$image_name:latest"
          fi
          echo "Build and push for $service_name complete."
          break
        fi
      done

      if [ "$service_found" = false ]; then
        services_not_found+=("$requested_service")
      fi
    done

    if [ ${#services_not_found[@]} -gt 0 ]; then
      echo "Error: The following services are not valid: ${services_not_found[*]}"
      # Dynamically list valid services
      valid_services=$(printf ", %s" "${SERVICES[@]}" | cut -c 3- | sed 's/:.*//g')
      echo "Valid services are: $valid_services"
      exit 1
    fi

    echo "Successfully built and pushed services: ${services_found[*]}"
  else
    # No specific services requested, so build and push all services.
    for service_config in "${SERVICES[@]}"; do
      IFS=':' read -r service_name service_dir image_name <<< "$service_config"

      echo "--- Building service: $service_name ---"

      # Go services need the services/ parent as build context so the shared
      # module (referenced via replace ../shared in go.mod) is accessible.
      local build_context="$service_dir"
      local -a build_file_args=()
      local -a build_arg_flags=()
      if [ "$service_name" != "ui" ]; then
        build_context="services"
        build_file_args=("-f" "$service_dir/Dockerfile")
      elif [ -n "$tag" ]; then
        build_arg_flags=("--build-arg" "VITE_APP_VERSION=$tag")
      fi
      if [ -n "$tag" ]; then
        echo "Building and pushing image: $image_name:$tag and $image_name:latest"
        docker build "${build_file_args[@]}" "${build_arg_flags[@]}" -t "$image_name:latest" -t "$image_name:$tag" "$build_context"
        docker push "$image_name:$tag"
        docker push "$image_name:latest"
      else
        echo "Building and pushing image: $image_name:latest"
        docker build "${build_file_args[@]}" "${build_arg_flags[@]}" -t "$image_name:latest" "$build_context"
        docker push "$image_name:latest"
      fi
      echo "Build and push for $service_name complete."
    done
  fi
}

# Deploy to Swarm
deploy_services() {
  local target_env="$1"
  local image_tag="$2"
  local deploy_compose_file=""
  local swarm_context=""
  local config_file=""
  local nginx_file=""
  local config_version_app=""
  local config_version_nginx=""
  local config_name=""
  local nginx_name=""
  local config_keep=""

  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  REPO_ROOT="$SCRIPT_DIR/.."

  require_image_tag "$image_tag"

  if [ "$target_env" == "qa" ]; then
    deploy_compose_file="$QA_COMPOSE_FILE"
    swarm_context="$QA_SWARM_CONTEXT"
    config_file="$REPO_ROOT/deployment/qa/partners-core.yml"
    nginx_file="$REPO_ROOT/deployment/nginx.conf"
  elif [ "$target_env" == "prod" ]; then
    deploy_compose_file="$PROD_COMPOSE_FILE"
    swarm_context="$PROD_SWARM_CONTEXT"
    config_file="$REPO_ROOT/deployment/prod/partners-core.yml"
    nginx_file="$REPO_ROOT/deployment/nginx.conf"
  else
    echo "Error: Missing or invalid deploy target. Use 'qa' or 'prod'."
    usage
  fi

  if [ ! -f "$config_file" ]; then
    echo "Error: Config file not found at $config_file"
    exit 1
  fi

  if [ ! -f "$nginx_file" ]; then
    echo "Error: Nginx config file not found at $nginx_file"
    exit 1
  fi

  config_keep="${CONFIG_KEEP:-$CONFIG_KEEP_DEFAULT}"
  config_version_app=$(compute_config_hash "$config_file")
  config_version_nginx=$(compute_config_hash "$nginx_file")
  config_name="partners-config-$config_version_app"
  nginx_name="partners-nginx-$config_version_nginx"

  # --- Swarm Deployment Logic (runs in all cases) ---
  echo "--- Deploying to Swarm ($target_env) ---"
    if ! docker context inspect "$swarm_context" &> /dev/null; then
      echo "Error: Docker context '$swarm_context' not found."
      exit 1
  fi

  echo "Switching to Docker context: $swarm_context"
  docker context use "$swarm_context"
  trap 'docker context use "$DEFAULT_CONTEXT"' EXIT

  require_registry_login_context "$swarm_context"
  require_images_exist "$swarm_context" "$image_tag"

  echo "Creating config objects (app: $config_version_app, nginx: $config_version_nginx)..."
  create_swarm_config "$swarm_context" "$config_name" "$config_file" "$target_env" "app" "$config_version_app"
  create_swarm_config "$swarm_context" "$nginx_name" "$nginx_file" "$target_env" "nginx" "$config_version_nginx"

  export CONFIG_VERSION_APP="$config_version_app"
  export CONFIG_VERSION_NGINX="$config_version_nginx"
  export IMAGE_TAG="$image_tag"

  echo "Deploying stack '$STACK_NAME' using $BASE_COMPOSE_FILE + $deploy_compose_file (image tag: $image_tag)..."
  docker stack deploy -c "$BASE_COMPOSE_FILE" -c "$deploy_compose_file" --with-registry-auth "$STACK_NAME"

  echo "Deployment command sent to Swarm."
  echo "Pruning old config objects (keep $config_keep per role)..."
  cleanup_swarm_configs "$swarm_context" "$target_env" "app" "$config_keep"
  cleanup_swarm_configs "$swarm_context" "$target_env" "nginx" "$config_keep"
  # registry_logout_context "$swarm_context"
  echo "Switching back to default Docker context: $DEFAULT_CONTEXT"
  docker context use "$DEFAULT_CONTEXT"
  trap - EXIT
  echo "--- Deployment process finished! ---"
}

migrate_database() {
  local target_env="$1"
  local direction="$2"  # up, down, force
  local service="$3"
  local version="$4"    # optional version number
  local db_name=""

  load_migrate_env "$target_env"

  if [ -z "$target_env" ] || [ -z "$direction" ]; then
    echo "Usage: $0 migrate <qa|prod> <up|down|force> [service] [version]"
    echo "Examples:"
    echo "  $0 migrate qa up auth 000001"
    echo "  $0 migrate qa down core 1"
    echo "  $0 migrate qa force core 2"
    exit 1
  fi

  if [ "$direction" != "up" ] && [ "$direction" != "down" ] && [ "$direction" != "force" ]; then
    echo "Error: Invalid direction '$direction'. Use 'up', 'down', or 'force'."
    exit 1
  fi

  declare -A service_db_map=(
    ["auth"]="$AUTH_SERVICE_DB_NAME"
    ["core"]="$CORE_SERVICE_DB_NAME"
    ["events"]="$EVENTS_SERVICE_DB_NAME"
    ["assets"]="$ASSET_SERVICE_DB_NAME"
    ["inventory"]="$INVENTORY_SERVICE_DB_NAME"
    ["sales-channel"]="$SALES_CHANNEL_SERVICE_DB_NAME"
  )

  if [ "$direction" == "down" ] || [ "$direction" == "force" ]; then
    if [ -z "$service" ] || [ -z "$version" ]; then
      echo "Error: '$direction' requires a service and version."
      echo "Example: $0 migrate qa $direction core 1"
      exit 1
    fi
  fi

  if [ -n "$service" ]; then
    if [[ ! " ${!service_db_map[@]} " =~ " $service " ]]; then
      echo "Error: Unknown service '$service'. Valid services are: ${!service_db_map[@]}"
      exit 1
    fi

    db_name="${service_db_map[$service]}"
    run_service_migration "$service" "$direction" "$version" "$db_name"
    echo "Migration completed for service '$service'."
    return
  fi

  if [ "$direction" == "down" ] || [ "$direction" == "force" ]; then
    echo "Error: '$direction' requires a service and version."
    exit 1
  fi

  for svc in "${MIGRATE_SERVICES[@]}"; do
    db_name="${service_db_map[$svc]}"
    run_service_migration "$svc" "$direction" "" "$db_name"
  done
  echo "Migration completed for all services."
}
# --- Main Logic ---
COMMAND=$1
if [ -z "$COMMAND" ]; then
  usage
fi

case "$COMMAND" in
  up)
    dev_up
    ;;
  down)
    dev_down
    ;;
  logs)
    dev_logs "$2"
    ;;
  build-push)
    shift # Remove "build-push" from arguments
    build_push_services "$@" # First arg is target env, remaining are services
    ;;
  deploy)
    shift # Remove "deploy" from arguments
    deploy_services "$@" # First arg is target env, remaining are services
    ;;
  migrate)
    shift
    migrate_database "$@"
    ;;
  logout)
    registry_logout
    ;;
  *)
    usage
    ;;
esac
