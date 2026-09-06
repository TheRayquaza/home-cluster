output "worker_details" {
  description = "Detailed map of worker node information"
  value = {
    for k, v in aws_instance.worker_node : k => {
      instance_id = v.id
      public_ip   = aws_eip.worker_eip[k].public_ip
      private_ip  = v.private_ip
      az          = v.availability_zone
    }
  }
}

output "worker_public_ips" {
  description = "List of all worker public IPs"
  value       = [for eip in aws_eip.worker_eip : eip.public_ip]
}
