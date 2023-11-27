cd config/crd/bases
for CRD in *.yaml; do
    if [ -f "../patches/${CRD}-patch" ]; then
        echo "Applying ${CRD}-patch"
        ${YAML_PATCH} -o "../patches/${CRD}-patch" < "${CRD}" > "${CRD}.patched"
        mv "${CRD}.patched" "${CRD}"
    fi
done