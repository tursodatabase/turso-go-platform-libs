#!/usr/bin/env bash
set -euo pipefail

TURSO_RS_REPO=${TURSO_RS_REPO:-https://github.com/tursodatabase/turso.git}
TURSO_RS_BUILD_PROFILE=${TURSO_RS_BUILD_PROFILE:-lib-release}
TURSO_RS_BUILD_DIR=${TURSO_RS_BUILD_DIR:-turso-rs}
TURSO_RS_PACKAGE=${TURSO_RS_PACKAGE:-turso_sync_sdk_kit}
TURSO_RS_LIBC_VARIANT=${TURSO_RS_LIBC_VARIANT:-""}
TURSO_GO_LIB_DIR=${TURSO_GO_LIB_DIR:-libs}

if [[ "$TURSO_RS_BUILD_REF" == "" ]]; then
    echo "TURSO_RS_BUILD_REF env var must be set"
    exit 1
fi

CARGO_ARGS_ARR=(--profile $TURSO_RS_BUILD_PROFILE)

UNAME_S="$(uname -s)"
UNAME_M="$(uname -m)"

echo "UNAME_S: $UNAME_S"
echo "UNAME_M: $UNAME_M"

case "$UNAME_S" in
  Linux*)  OS=linux  ;;
  Darwin*) OS=darwin ;;
  MINGW*|MSYS*|CYGWIN*) OS=windows ;;
  *) echo "Unsupported OS: $UNAME_S"; exit 1 ;;
esac

if [[ "$OS" == "windows" ]]; then
  case "$UNAME_S" in
    *ARM64)  ARCH=arm64  ;;
    *) ARCH=amd64 ;;
  esac
else
  # cygwin reports x86_64 even for windows on ARM
  case "$UNAME_M" in
    x86_64) ARCH=amd64 ;;
    arm64|aarch64) ARCH=arm64 ;;
    i386|i686) ARCH=386 ;;
    *) echo "Unsupported arch: $UNAME_M"; exit 1 ;;
  esac
fi

# Detect libc variant on Linux
if [[ "$TURSO_RS_LIBC_VARIANT" == "" ]];
then
    if [[ "$UNAME_S" == Linux* ]]; then
        # Check if we're on musl-based system (Alpine)
        if ldd --version 2>&1 | grep -qi musl; then
            TURSO_RS_LIBC_VARIANT="_musl"
        elif [ -f /etc/alpine-release ]; then
            # Fallback detection for Alpine Linux
            TURSO_RS_LIBC_VARIANT="_musl"
        fi
    fi
fi

PLATFORM="${OS}_${ARCH}${TURSO_RS_LIBC_VARIANT}"
TURSO_GO_LIB_PATH="${TURSO_GO_LIB_DIR}/${PLATFORM}"

case "$OS" in
  linux)
    if [[ "$TURSO_RS_LIBC_VARIANT" == "_musl" ]]; then
      OUTPUT_NAME="lib${TURSO_RS_PACKAGE}.a"
    else
      OUTPUT_NAME="lib${TURSO_RS_PACKAGE}.so"
    fi
    ;;
  darwin)  OUTPUT_NAME="lib${TURSO_RS_PACKAGE}.dylib" ;;
  windows) OUTPUT_NAME="${TURSO_RS_PACKAGE}.dll" ;;
esac

# Set Rust target for musl builds
RUST_TARGET=""
if [[ "$TURSO_RS_LIBC_VARIANT" == "_musl" ]]; then
  case "$ARCH" in
    amd64) RUST_TARGET="x86_64-unknown-linux-musl" ;;
    arm64) RUST_TARGET="aarch64-unknown-linux-musl" ;;
    *) echo "Unsupported musl arch: $ARCH"; exit 1 ;;
  esac
  # Check if rustup is available and if the target needs to be installed
  if command -v rustup >/dev/null 2>&1; then
    if ! rustup target list --installed | grep -q "$RUST_TARGET"; then
      echo "Installing Rust target: $RUST_TARGET"
      rustup target add "$RUST_TARGET"
    fi
  else
    # rustup not available (e.g., Rust installed via package manager)
    # Assume the musl target is already available or will be handled by cargo
    echo "rustup not found, assuming $RUST_TARGET is available"
  fi
  CARGO_ARGS_ARR+=("--target" "$RUST_TARGET")
fi

# Determine output directory (changes when using --target)
if [[ -n "$RUST_TARGET" ]]; then
  CARGO_OUT_DIR="${TURSO_RS_BUILD_DIR}/target/${RUST_TARGET}/${TURSO_RS_BUILD_PROFILE}"
else
  CARGO_OUT_DIR="${TURSO_RS_BUILD_DIR}/target/${TURSO_RS_BUILD_PROFILE}"
fi
CARGO_LIB_PATH="${CARGO_OUT_DIR}/${OUTPUT_NAME}"

echo "TURSO_RS_REPO: $TURSO_RS_REPO"
echo "TURSO_RS_BUILD_REF: $TURSO_RS_BUILD_REF"
echo "TURSO_RS_BUILD_DIR: $TURSO_RS_BUILD_DIR"
echo "TURSO_RS_PACKAGE: $TURSO_RS_PACKAGE"
echo "CARGO_ARGS_ARR: ${CARGO_ARGS_ARR[@]}"
echo "CARGO_OUT_DIR: $CARGO_OUT_DIR"
echo "CARGO_LIB_PATH: $CARGO_LIB_PATH"
echo "TURSO_GO_LIB_PATH: $TURSO_GO_LIB_PATH"

git clone --single-branch --depth 1 --branch $TURSO_RS_BUILD_REF $TURSO_RS_REPO $TURSO_RS_BUILD_DIR

pushd $TURSO_RS_BUILD_DIR
echo "Building ${TURSO_RS_PACKAGE} ($TURSO_RS_BUILD_PROFILE) for ${PLATFORM}"
export CARGO_ARGS="${CARGO_ARGS_ARR[@]}"
cargo build "${CARGO_ARGS_ARR[@]}" --package "${TURSO_RS_PACKAGE}"
popd


if [[ ! -f "$CARGO_LIB_PATH" ]]; then
  echo "Expected artifact not found: $CARGO_LIB_PATH"
  echo "Contents of ${CARGO_OUT_DIR}:"
  ls -la "${CARGO_OUT_DIR}" || true
  exit 1
fi

mkdir -p "${TURSO_GO_LIB_PATH}"
cp -f "${CARGO_LIB_PATH}" "${TURSO_GO_LIB_PATH}/"

echo "Wrote $(pwd)/${TURSO_GO_LIB_PATH}/$(basename $CARGO_LIB_PATH)"