output "external_ip" {
  value       = yandex_vpc_address.static-ip.external_ipv4_address[0].address
  description = "Статический публичный IP-адрес виртуальной машины"
}
