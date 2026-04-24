variable "cloud_id" {
  description = "ID облака Yandex Cloud"
  type        = string
  default     = "b1gm71hm7rugdu58n8ij"
}

variable "folder_id" {
  description = "ID каталога Yandex Cloud"
  type        = string
  default     = "b1g817dsjhr6fr9hn2uf"
}

variable "zone" {
  description = "Зона доступности"
  type        = string
  default     = "ru-central1-a"
}

variable "sa_key_file" {
  description = "Путь к ключу сервисного аккаунта"
  type        = string
  default     = "/home/radima1/.terraform/authorized_key.json"
}

variable "public_key_path" {
  description = "Путь к публичному SSH-ключу"
  type        = string
  default     = "/home/radima1/.ssh/id_rsa.pub"
}

variable "vm_name" {
  description = "Имя виртуальной машины"
  type        = string
  default     = "sitechecker-vm"
}

variable "db_password" {
  description = "Пароль для PostgreSQL"
  type        = string
  sensitive   = true
}
