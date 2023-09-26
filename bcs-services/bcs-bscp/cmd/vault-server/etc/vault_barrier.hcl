ui = false
api_addr = "http://${POD_IP}:8200"
disable_mlock = true

storage "mysql"{
  address  = "127.0.0.1"
  username = "root"
  password = ""
  database = "vault"
  ha_enabled = true
  plaintext_connection_allowed = true
}

listener "tcp"{
  address = "${POD_IP}:8200"

  tls_disable = true
}