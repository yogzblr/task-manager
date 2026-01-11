#!/bin/bash

# Submit 10 jobs, each echoing its number
for i in {1..10}
do
  echo "Submitting job $i..."
  curl -X POST http://localhost:8081/api/jobs \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhZ2VudC1saW51eC0wMSIsImFnZW50X2lkIjoiYWdlbnQtbGludXgtMDEiLCJ0ZW5hbnRfaWQiOiJ0ZXN0LXRlbmFudCIsInByb2plY3RfaWQiOiJ0ZXN0LXByb2plY3QiLCJjaGFubmVscyI6WyJhZ2VudHMudGVzdC10ZW5hbnQuYWdlbnQtbGludXgtMDEiXSwiZXhwIjoxNzk5NjQ4MTgxLCJpYXQiOjE3NjgxMTIxODF9.QwRFU1QEiszhgl5YKC2kt88Ai7az1_haPVAvmi2acHU" \
    -d "{
      \"tenant_id\": \"test-tenant\",
      \"project_id\": \"test-project\",
      \"agent_id\": \"agent-linux-01\",
      \"workflow\": \"tasks:\n  - name: job-$i\n    type: command\n    config:\n      command: echo\n      args:\n        - \\\"$i\\\"\n\"
    }"
  echo ""
  sleep 0.5
done

echo "All 10 jobs submitted!"
