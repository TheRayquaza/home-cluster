resource "aws_instance" "worker_node" {
  ami                    = data.aws_ami.ubuntu.id
  for_each               = var.worker_nodes
  availability_zone      = each.value.az
  subnet_id              = aws_subnet.public[each.key].id
  vpc_security_group_ids = [aws_security_group.wg_sg.id]
  key_name               = each.value.key_name
  instance_type          = each.value.instance_type

  source_dest_check = false

  root_block_device {
    volume_type           = "gp3"
    volume_size           = 30
    delete_on_termination = false
  }

  tags = {
    Name = "K8S-${each.key}"
  }
}
