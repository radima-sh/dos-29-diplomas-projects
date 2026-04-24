output "external_ip" {
  value       = yandex_compute_instance.sitechecker-vm.network_interface.0.nat_ip_address
  description = "Публичный IP-адрес созданной виртуальной машины"
}
