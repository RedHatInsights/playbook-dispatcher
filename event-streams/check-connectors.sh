#!/usr/bin/env bash

# https://rmoff.net/2019/06/06/automatically-restarting-failed-kafka-connect-tasks/

curl -s "http://${CONNECT_HOST:-localhost}:${CONNECT_PORT:-8083}/connectors?expand=status" | \
  jq -c -M 'map({name: .status.name } +  {tasks: .status.tasks}) | .[] | {task: ((.tasks[]) + {name: .name})}  | select(.task.state=="FAILED") | {name: .task.name, task_id: .task.id|tostring} | ("/connectors/"+ .name + "/tasks/" + .task_id + "/restart")' | \
  xargs -I{connector_and_task} curl -v -X POST "http://${CONNECT_HOST:-localhost}:${CONNECT_PORT:-8083}"\{connector_and_task\}
