docker buildx build \
  --platform linux/arm64 \
  --tag $IMAGE \
  --push \
  $BUILD_CONTEXT