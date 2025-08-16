id_value=$(jq -r '.order_uid' model.json)
curl "http://localhost:8080/order?id=$id_value"