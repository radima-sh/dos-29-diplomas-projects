# 1. Виртуальная сеть (VPC)
resource "yandex_vpc_network" "sitechecker-net" {
  name = "sitechecker-network"
}

# 2. Подсеть (в зоне ru-central1-a)
resource "yandex_vpc_subnet" "sitechecker-subnet" {
  name           = "sitechecker-subnet-a"
  zone           = var.zone
  network_id     = yandex_vpc_network.sitechecker-net.id
  v4_cidr_blocks = ["10.0.1.0/24"]
}

# 3. Группа безопасности (облачный фаервол)
resource "yandex_vpc_security_group" "sitechecker-sg" {
  name       = "sitechecker-security-group"
  network_id = yandex_vpc_network.sitechecker-net.id

  # Разрешаем входящий SSH (порт 22)
  ingress {
    protocol          = "TCP"
    description       = "Allow SSH"
    port              = 22
    predefined_target = "self_security_group"
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }

  # Разрешаем входящий HTTP (порт 80)
  ingress {
    protocol          = "TCP"
    description       = "Allow HTTP"
    port              = 80
    predefined_target = "self_security_group"
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }

  # Разрешаем входящий HTTPS (порт 443)
  ingress {
    protocol          = "TCP"
    description       = "Allow HTTPS"
    port              = 443
    predefined_target = "self_security_group"
    v4_cidr_blocks    = ["0.0.0.0/0"]
  }

  # Разрешаем весь исходящий трафик
  egress {
    protocol       = "TCP"
    description    = "Allow all outbound"
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
}

# 4. Поиск актуального образа Ubuntu 22.04 LTS
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-2204-lts"
}

# 5. Виртуальная машина (Compute Instance)
resource "yandex_compute_instance" "sitechecker-vm" {
  name        = var.vm_name
  platform_id = "standard-v3" # Современная платформа (Intel Ice Lake)

  # Ресурсы ВМ
  resources {
    cores  = 2
    memory = 4
  }

  # Загрузочный диск
  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu.id
      size     = 30 # 30 ГБ SSD
    }
  }

  # Сетевой интерфейс
  network_interface {
    subnet_id          = yandex_vpc_subnet.sitechecker-subnet.id
    nat                = true # true = выдать публичный IP-адрес
    security_group_ids = [yandex_vpc_security_group.sitechecker-sg.id]
  }

  # Метаданные: вставляем публичный SSH-ключ для пользователя ubuntu
  metadata = {
    ssh-keys = "ubuntu:${file(var.public_key_path)}"
  }
}
