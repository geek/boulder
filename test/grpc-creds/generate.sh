#!/bin/bash
set -e
set -o xtrace

# Check that `minica` is installed
command -v minica >/dev/null 2>&1 || {
  echo >&2 "No 'minica' command available.";
  echo >&2 "Check your GOPATH and run: 'go get github.com/jsha/minica'.";
  exit 1;
}

for SERVICE in admin-revoker expiration-mailer ocsp-updater ocsp-responder \
  orphan-finder wfe akamai-purger bad-key-revoker crl-updater crl-storer \
  health-checker; do
  minica -domains "${SERVICE}.boulder"
done

NEEDIPSANS=( "sa" )
for SERVICE in publisher nonce ra ca sa va ; do
  if [[ "${NEEDIPSANS[@]}" =~ "${SERVICE}" ]]; then
    minica -domains "${SERVICE}.boulder,${SERVICE}1.boulder,${SERVICE}2.boulder" \
      -ip-addresses "10.77.77.77,10.88.88.88"
  else
    minica -domains "${SERVICE}.boulder,${SERVICE}1.boulder,${SERVICE}2.boulder"
  fi
done
