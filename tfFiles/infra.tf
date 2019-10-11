provider "aws" {
  profile = "default"
  region = "us-east-2"
}

resource "aws_vpc" "hpcfl_vpc" {
  cidr_block = "192.168.2.0/24"

  tags = {
    Name = "hpcfl_vpc"
  }
}

resource "aws_internet_gateway" "hpcfl_gw" {
  vpc_id = "${aws_vpc.hpcfl_vpc.id}"

  tags = {
    Name = "hpcfl_gw"
  }
}

resource "aws_route_table" "hpcfl_route_table" {
  vpc_id = "${aws_vpc.hpcfl_vpc.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.hpcfl_gw.id}"
  }
  tags = {
    Name = "hpcfl_route_table"
    Main = "Yes"
  }
}

resource "aws_main_route_table_association" "hpcfl_rt_a" {
  vpc_id = "${aws_vpc.hpcfl_vpc.id}"
  route_table_id = "${aws_route_table.hpcfl_route_table.id}"
}

resource "aws_security_group" "hpcfl_wireguard" {
  name = "hpcfl_wireguard"
  vpc_id = "${aws_vpc.hpcfl_vpc.id}"
  ingress {
    from_port = 22
    to_port   = 22
    protocol = "tcp"
    cidr_blocks = ["128.117.0.0/16"]
  }
  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  #ingress {
  #  from_port =  33434
  #  to_port = 33434
  #  protocol = "udp"
  #  cidr_blocks = ["128.117.0.0/16"]
  #}
}

resource "aws_subnet" "hpcfl_subnet" {
  vpc_id = "${aws_vpc.hpcfl_vpc.id}"
  cidr_block = "192.168.2.0/24"
  availability_zone = "us-east-2a"
}

resource "aws_key_pair" "hpcfl2" {
  key_name = "hpcfl2"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC0LOWv1UH0VRoYbtnVbhOki2f229ObuDMaxcOEpmYe4Ckk3oT7fAjUijNtb2eWJP0jAQlzmVRj4Rd9K1R1ufrY+7sIp40/0x2zzDqRNxW4AMi7mfAS0aF/i60y5zSZBgicDfmyfgjx+f3GBEJLuwo5eUB2cZ6P8l5XI+rzisiwTCAN5J9l4z4o67uxVj9kJq1lZIW13X0r+ZmTeRpIHZgOs6EUD3g1Bh5dwBayglcansLXjtqFOe6jO0TV1ou0rZT70YxfI2dcFgA8fTvBEAUvx7eT8vaQ2WQWa/mdx1+pHXinAtwAvuNleB2fnhMV45W68RfD3O0a82ZCZpSxYown shanks@hlwill"
}
