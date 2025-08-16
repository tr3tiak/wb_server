id_value=$(jq -r '.order_uid' model.json)
curl -w "Response Time: %{time_total}s, Speed: %{speed_download} bytes/sec\n" \
  "http://localhost:8080/order?id=$id_value" -o /dev/null -s