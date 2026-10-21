#!/usr/bin/env bash
#
# Mitigation for the eMudhra intermediate CA private key compromise
# (RHSB-2026-012 / CVE-2026-111943).
#
# The compromised intermediates chain to two root CAs that ship in the
# RHEL/UBI system trust store:
#   - "Comodo AAA Services root" (aka "AAA Certificate Services")
#   - "emSign ECC Root CA - G3"
#
# Until the base image's ca-certificates package is updated to drop trust
# in these roots, explicitly distrust them via the p11-kit blocklist. This
# mirrors the steps in:
#   https://access.redhat.com/solutions/7149312
#   https://access.redhat.com/security/vulnerabilities/RHSB-2026-012
#
# Safe to run multiple times and safe to keep running after the base image
# ships a fix: `trust extract` simply writes an empty file when the
# filtered certificate is no longer present in the trust store.

set -euo pipefail

if ! command -v trust >/dev/null 2>&1; then
    echo "disable-compromised-ca-certs: 'trust' (p11-kit-trust) not found; skipping" >&2
    exit 0
fi

blocklist_dir=/etc/pki/ca-trust/source/blocklist
[ -d "$blocklist_dir" ] || blocklist_dir=/etc/pki/ca-trust/source/blacklist
mkdir -p "$blocklist_dir"

trust extract --overwrite --format=pem-bundle \
    --filter='pkcs11:object=Comodo%20AAA%20Services%20root;type=cert' \
    "$blocklist_dir/Comodo_AAA_Services_root.pem"

trust extract --overwrite --format=pem-bundle \
    --filter='pkcs11:object=emSign%20ECC%20Root%20CA%20-%20G3;type=cert' \
    "$blocklist_dir/emSign_ECC_Root_CA_-_G3.pem"

update-ca-trust extract

echo "disable-compromised-ca-certs: distrusted 'Comodo AAA Services root' and 'emSign ECC Root CA - G3'"
