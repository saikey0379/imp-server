## 依赖
- MySQL（5.6+）
- Go1.21及以上版本
- Docker / Kubernetes (可选)

## 安装
### Linux下安装编译环境
1. 登录[golang官网](https://golang.org/dl/)或者[golang中国官方镜像](https://golang.google.cn/dl/)下载最新的稳定版本的go安装包并安装。

	```bash
	$ wget https://go.dev/dl/go1.21.1.linux-amd64.tar.gz
	# 解压缩后go被安装在/usr/local/go
	$ sudo tar -xzvf ./go1.21.1.linux-amd64.tar.gz -C /usr/local/
	```

2. 配置go环境变量

	```bash
	$ cat << "EOF" >> ~/.bashrc
	export GOROOT=/usr/local/go
	export PATH=$PATH:$GOROOT/bin
	EOF
	$ source ~/.bashrc
	```
 
### 源代码编译
1. 源码下载与编译

	```bash
	$ git clone https://github.com/saikey0379/imp-server.git
	
	$ cd /imp-server
	$ go build -o ./bin/imp-server ./cmd/main.go
	```
 
	```bash
	$ ls -l bin
	total 133848
	-rwxr-xr-x  1 root  root    16M  3  1 10:36 imp-server
	```
2. 安装

	```bash
	$ IMP_INSTALL_DIR=/usr/local/imp/
	$ install -p -m 0755 bin/imp-server ${IMP_INSTALL_DIR}/bin/
	$ install -p -m 0644 conf/imp-server.conf ${IMP_INSTALL_DIR}/conf/

	$ /usr/local/imp/bin/imp-server -v
	imp-server version 0.0.1
	```
3. RPMbuild(可选)

	```bash
	$ yum -y install rpmbuild
	
	$ VERSION=v0.0.1
	$ tar -zcvf /root/rpmbuild/SOURCES/imp-server-${VERSION}.tgz bin/ conf/ deploy/systemd/
	$ sed "s/VERSION/${VERSION}/g" deploy/rpmbuild/imp-server.spec > imp-server_${VERSION}.spec
	$ rpmbuild -bb imp-server_${VERSION}.spec
	$ mv /root/rpmbuild/RPMS/x86_64/imp-server-${VERSION}-0.x86_64.rpm .
	```

### 初始化数据
1. 导入SQL文件初始化数据库
将`./doc/db/imp-server.sql`导入MySQL。

2. 配置文件修改`/usr/local/imp/conf/imp-server.conf`

	```ini
	[Server]
	listen = "0.0.0.0"
	port = 8083
	redisAddr = "127.0.0.1"
	redisPort = 6379
	redisPasswd = "password"
	redisDBNumber = 1

	[Pxe]
	pxeConfigDir = "/var/lib/tftpboot/pxelinux.cfg"

	[Repo]
	connection = "root:imp@tcp(127.0.0.1:3306)/imp?charset=utf8&parseTime=True&loc=Local"

	[Logger]
	logFile = "/var/log/imp-server.log"
	level = "debug"
	```

## 启动
### 裸机启动(可选)
##### 推荐systemd启动
1. 服务启动
	```bash
	$ /usr/local/imp/bin/imp-server -c /usr/local/imp/conf/imp-server.conf
	```
2. Systemd启动
   ##### 通过rpm包安装可以跳过此步
	```bash
	$ install -p -m 0755 deploy/systemd/imp-server.service /lib/systemd/system/imp-server.service
 	```
   ##### 启动
	```bash
	$ systemctl enable imp-server && systemctl start imp-server
	```
### 容器启动(可选)
1. 镜像构建
	```bash
	$ docker build -t docker.example.com/imp-server:v0.0.1 -f Dockerfile .
	```
2. 容器启动
	```bash
	# 容器环境配置文件调整可能较为频繁，建议挂载外部配置
	$ install -p -m 0644 conf/imp-server.conf /usr/local/imp/conf/

	$ docker run -d -v /var/lib/tftpboot/pxelinux.cfg/:/var/lib/tftpboot/pxelinux.cfg/ -v /usr/local/imp/conf/imp-server.conf:/usr/local/imp/conf/imp-server.conf imp-server:v0.0.1 --name imp-server
 	```
### Kubernetes启动(可选)
##### 镜像构建【同上】
1. 启动
	```bash
	$ kubectl apply -f doc/kubernetes/
	$ kubectl get all
 	```
   
## 附.基础服务安装参考
### HTTP安装
1. 安装
	```bash
	$ rpm -Uvh http://nginx.org/packages/centos/7/noarch/RPMS/nginx-release-centos-7-0.el7.ngx.noarch.rpm
	$ yum -y install nginx
	```
2. 配置
	```bash
	$ cat << "EOF" > /etc/nginx/conf.d/imp-server.conf 
	server {
		listen 80;
		server_name imp.example.com;
 		# imp-server服务api
		location /api/ {
			proxy_pass http://127.0.0.1:8083;
		}
 
		# 存放os镜像以及安装包
    	location /www {
			root   /data/nginx/;
			autoindex on;
			autoindex_exact_size off;
			autoindex_localtime on;
    	}
	}
	EOF
	$ systemctl enable nginx && systemctl start nginx
	```
### DNS安装
##### 如已存在DNS服务器则跳过第一步

1. 安装Dnsmasq
	```bash
	$ yum -y install dnsmasq
	$ systemctl enable dnsmasqd && systemctl start dnsmasqd
	```
 2. 获取服务访问地址，并配置DNS服务器
	```bash
	$ IPADDR=$(ifconfig $(route -n | grep ^0.0.0.0 | awk '{print$NF}') | grep netmask | head -n 1 | awk '{print$2}')
	$ cat << EOF > /etc/dnsmasqd.conf
	address=/imp.example.com/${IPADDR}
	EOF
	$ systemctl restart dnsmasqd
	```
### DHCP安装
1. 安装配置
	```bash
	# 替换成真实的地址
	$ SERVER_DNS=192.168.0.1
	$ SERVER_TFTP=$(ifconfig $(route -n | grep ^0.0.0.0 | awk '{print$NF}') | grep netmask | head -n 1 | awk '{print$2}')

	$ yum -y install dhcp
	$ cat << EOF > /etc/dhcp/dhcpd.conf
	allow booting;
	allow bootp;
	ddns-update-style none;
	ping-check true;
	ping-timeout 3;
	default-lease-time 120;
	max-lease-time 600;
	authoritative;
	filename "undionly.kkpxe";
	next-server ${SERVER_TFTP};
	option domain-name-servers ${SERVER_DNS};
 
	subnet 192.168.0.0 netmask 255.255.255.0 {
        range 192.168.0.2 192.168.0.254;
	}
	EOF
	$ systemctl enable dhcpd && systemctl start dhcpd
	```
### TFTP安装
1. 安装配置

	```bash
	$ yum -y install tftp-server

	# service tftp
	$ sed -i "/disable/ s/yes/no/g" /etc/xinetd.d/tftp
 
	$ cat << "EOF" > /var/lib/tftpboot/pxelinux.cfg/default 
	DEFAULT IMPOS
	LABEL IMPOS
	KERNEL http://imp.example.com/www/os/centos/7.9/images/pxeboot/vmlinuz
	APPEND initrd=http://imp.example.com/www/os/IMPOS.img
	IPAPPEND 2
	EOF
 
	$ systemctl enable tftp && systemctl start tftp
	```
