#!/bin/bash
IFS=', ' read -r -a addr_array <<<"$SERVER_ADDRR"
containerName="${CI_PROJECT_NAME}"

privateCertPath=/home/gitlab-runner/
publicCertPath=/home/gitlab-runner/

for addr in "${addr_array[@]}"; do
  SSH_CMD="ssh -o StrictHostKeyChecking=no gitlab-runner@${addr}"

  $SSH_CMD docker login "${CI_REGISTRY}" -u gitlab-ci-token -p "${CI_JOB_TOKEN}"
  $SSH_CMD docker pull "${IMAGE}" || exit 1
  $SSH_CMD docker stop "${containerName}" || echo "nothing to stop"
  $SSH_CMD docker rm "${containerName}" || echo "nothing to remove"
  $SSH_CMD docker run -d --net=host --name "${containerName}" -h "\$HOSTNAME" \
    --restart=always \
    --log-opt max-size=200m --log-opt max-file=3 ${IMAGE} \
    -v /opt/ironmaiden/config:/opt/ironmaiden/config \

  $SSH_CMD docker ps
  $SSH_CMD docker logs "${containerName}"
done
