package template

import (
	"github.com/minectl/pgk/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	bedrock = model.MinecraftServer{
		Spec: model.Spec{
			Minecraft: model.Minecraft{
				Edition:    "bedrock",
				Properties: "level-seed=stackitminecraftrocks\nview-distance=10\nenable-jmx-monitoring=false\n",
			},
		},
	}
	java = model.MinecraftServer{
		Spec: model.Spec{
			Minecraft: model.Minecraft{
				Java: model.Java{
					Xms: "2G",
					Xmx: "2G",
					Rcon: model.Rcon{
						Port:      2,
						Password:  "test",
						Enabled:   true,
						Broadcast: true,
					},
				},
				Edition:    "java",
				Properties: "level-seed=stackitminecraftrocks\nview-distance=10\nenable-jmx-monitoring=false\n",
			},
		},
	}
	bedrockCivoWant = `#!/bin/bash

tee /tmp/server.properties <<EOF
level-seed=stackitminecraftrocks
view-distance=10
enable-jmx-monitoring=false

EOF

tee /tmp/prometheus.yml <<EOF
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'prometheus'
    scrape_interval: 5s
    static_configs:
      - targets: ['localhost:9090']
  - job_name: 'node_exporter'
    scrape_interval: 5s
    static_configs:
      - targets: ['localhost:9100']
EOF
tee /etc/systemd/system/prometheus.service <<EOF
[Unit]
Description=Prometheus
Wants=network-online.target
After=network-online.target

[Service]
User=prometheus
Group=prometheus
Type=simple
ExecStart=/usr/local/bin/prometheus \
    --config.file /etc/prometheus/prometheus.yml \
    --storage.tsdb.path /var/lib/prometheus/ \
    --web.console.templates=/etc/prometheus/consoles \
    --web.console.libraries=/etc/prometheus/console_libraries

[Install]
WantedBy=multi-user.target
EOF
tee /etc/systemd/system/node_exporter.service <<EOF
[Unit]
Description=Node Exporter
Wants=network-online.target
After=network-online.target

[Service]
User=node_exporter
Group=node_exporter
Type=simple
ExecStart=/usr/local/bin/node_exporter

[Install]
WantedBy=multi-user.target
EOF
tee /etc/systemd/system/minecraft.service <<EOF
[Unit]
Description=Minecraft Server
Documentation=https://www.minecraft.net/en-us/download/server

[Service]
WorkingDirectory=/minecraft
Type=simple
ExecStart=/bin/sh -c "LD_LIBRARY_PATH=. ./bedrock_server"
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
apt-get install -y apt-transport-https ca-certificates curl unzip fail2ban

useradd prometheus -s /bin/false
useradd node_exporter -s /bin/false

export PROM_VERSION=2.28.1
mkdir /etc/prometheus
mkdir /var/lib/prometheus
curl -sSL https://github.com/prometheus/prometheus/releases/download/v$PROM_VERSION/prometheus-$PROM_VERSION.linux-amd64.tar.gz | tar -xz
cp prometheus-$PROM_VERSION.linux-amd64/prometheus /usr/local/bin/
cp prometheus-$PROM_VERSION.linux-amd64/promtool /usr/local/bin/
chown prometheus:prometheus /usr/local/bin/prometheus
chown prometheus:prometheus /usr/local/bin/promtool
cp -r prometheus-$PROM_VERSION.linux-amd64/consoles /etc/prometheus
cp -r prometheus-$PROM_VERSION.linux-amd64/console_libraries /etc/prometheus
chown -R prometheus:prometheus /var/lib/prometheus
chown -R prometheus:prometheus /etc/prometheus/consoles
chown -R prometheus:prometheus /etc/prometheus/console_libraries
mv /tmp/prometheus.yml /etc/prometheus/prometheus.yml
chown prometheus:prometheus /etc/prometheus/prometheus.yml
systemctl daemon-reload
systemctl start prometheus
systemctl enable prometheus

export NODE_EXPORTER_VERSION=1.1.2
curl -sSL https://github.com/prometheus/node_exporter/releases/download/v$NODE_EXPORTER_VERSION/node_exporter-$NODE_EXPORTER_VERSION.linux-amd64.tar.gz | tar -xz
cp node_exporter-$NODE_EXPORTER_VERSION.linux-amd64/node_exporter /usr/local/bin
chown node_exporter:node_exporter /usr/local/bin/node_exporter
systemctl daemon-reload
systemctl start node_exporter
systemctl enable node_exporter

ufw allow ssh
ufw allow 5201
ufw allow proto udp to 0.0.0.0/0 port 19132

echo [DEFAULT] | sudo tee -a /etc/fail2ban/jail.local
echo banaction = ufw | sudo tee -a /etc/fail2ban/jail.local
echo [sshd] | sudo tee -a /etc/fail2ban/jail.local
echo enabled = true | sudo tee -a /etc/fail2ban/jail.local
sudo systemctl restart fail2ban
mkdir /minecraft
curl -sLSf https://minecraft.azureedge.net/bin-linux/bedrock-server-1.17.2.01.zip > /tmp/bedrock-server.zip
unzip -o /tmp/bedrock-server.zip -d /minecraft
chmod +x /minecraft/bedrock_server
echo "eula=true" > /minecraft/eula.txt
mv /tmp/server.properties /minecraft/server.properties
systemctl restart minecraft.service
systemctl enable minecraft.service`

	javaCivoWant = `#!/bin/bash

tee /tmp/server.properties <<EOF
level-seed=stackitminecraftrocks
view-distance=10
enable-jmx-monitoring=false
broadcast-rcon-to-ops=true
rcon.port=2
enable-rcon=true
rcon.password=test
EOF

tee /tmp/prometheus.yml <<EOF
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'prometheus'
    scrape_interval: 5s
    static_configs:
      - targets: ['localhost:9090']
  - job_name: 'node_exporter'
    scrape_interval: 5s
    static_configs:
      - targets: ['localhost:9100']
  - job_name: 'minecraft_exporter'
    scrape_interval: 1m
    static_configs:
      - targets: ['localhost:9150']
EOF
tee /etc/systemd/system/prometheus.service <<EOF
[Unit]
Description=Prometheus
Wants=network-online.target
After=network-online.target

[Service]
User=prometheus
Group=prometheus
Type=simple
ExecStart=/usr/local/bin/prometheus \
    --config.file /etc/prometheus/prometheus.yml \
    --storage.tsdb.path /var/lib/prometheus/ \
    --web.console.templates=/etc/prometheus/consoles \
    --web.console.libraries=/etc/prometheus/console_libraries

[Install]
WantedBy=multi-user.target
EOF
tee /etc/systemd/system/node_exporter.service <<EOF
[Unit]
Description=Node Exporter
Wants=network-online.target
After=network-online.target

[Service]
User=node_exporter
Group=node_exporter
Type=simple
ExecStart=/usr/local/bin/node_exporter

[Install]
WantedBy=multi-user.target
EOF
tee /etc/systemd/system/minecraft-prometheus-exporter.service <<EOF
[Unit]
Description=Minecraft Exporter
Wants=network-online.target
After=network-online.target
[Service]
User=minecraft_exporter
Group=minecraft_exporter
Type=simple
ExecStart=/usr/local/bin/minecraft-prometheus-exporter \
  --mc.rcon-password=test
[Install]
WantedBy=multi-user.target
EOF

tee /etc/systemd/system/minecraft.service <<EOF
[Unit]
Description=Minecraft Server
Documentation=https://www.minecraft.net/en-us/download/server

[Service]
WorkingDirectory=/minecraft
Type=simple
ExecStart=/usr/bin/java -Xmx2G -Xms2G -jar server.jar nogui

Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
apt-get install -y apt-transport-https ca-certificates curl openjdk-16-jre-headless fail2ban

useradd prometheus -s /bin/false
useradd node_exporter -s /bin/false
useradd minecraft_exporter -s /bin/false

export PROM_VERSION=2.28.1
mkdir /etc/prometheus
mkdir /var/lib/prometheus
curl -sSL https://github.com/prometheus/prometheus/releases/download/v$PROM_VERSION/prometheus-$PROM_VERSION.linux-amd64.tar.gz | tar -xz
cp prometheus-$PROM_VERSION.linux-amd64/prometheus /usr/local/bin/
cp prometheus-$PROM_VERSION.linux-amd64/promtool /usr/local/bin/
chown prometheus:prometheus /usr/local/bin/prometheus
chown prometheus:prometheus /usr/local/bin/promtool
cp -r prometheus-$PROM_VERSION.linux-amd64/consoles /etc/prometheus
cp -r prometheus-$PROM_VERSION.linux-amd64/console_libraries /etc/prometheus
chown -R prometheus:prometheus /var/lib/prometheus
chown -R prometheus:prometheus /etc/prometheus/consoles
chown -R prometheus:prometheus /etc/prometheus/console_libraries
mv /tmp/prometheus.yml /etc/prometheus/prometheus.yml
chown prometheus:prometheus /etc/prometheus/prometheus.yml
systemctl daemon-reload
systemctl start prometheus
systemctl enable prometheus

export NODE_EXPORTER_VERSION=1.1.2
curl -sSL https://github.com/prometheus/node_exporter/releases/download/v$NODE_EXPORTER_VERSION/node_exporter-$NODE_EXPORTER_VERSION.linux-amd64.tar.gz | tar -xz
cp node_exporter-$NODE_EXPORTER_VERSION.linux-amd64/node_exporter /usr/local/bin
chown node_exporter:node_exporter /usr/local/bin/node_exporter
systemctl daemon-reload
systemctl start node_exporter
systemctl enable node_exporter

export MINECRAFT_EXPORTER_VERSION=0.3.1
curl -sSL https://github.com/dirien/minecraft-prometheus-exporter/releases/download/v$MINECRAFT_EXPORTER_VERSION/minecraft-prometheus-exporter_$MINECRAFT_EXPORTER_VERSION.linux-amd64.tar.gz | tar -xz
cp minecraft-prometheus-exporter /usr/local/bin
chown minecraft_exporter:minecraft_exporter /usr/local/bin/minecraft-prometheus-exporter
systemctl start minecraft-prometheus-exporter.service
systemctl enable minecraft-prometheus-exporter.service
ufw allow ssh
ufw allow 5201

ufw allow proto udp to 0.0.0.0/0 port 25565


echo [DEFAULT] | sudo tee -a /etc/fail2ban/jail.local
echo banaction = ufw | sudo tee -a /etc/fail2ban/jail.local
echo [sshd] | sudo tee -a /etc/fail2ban/jail.local
echo enabled = true | sudo tee -a /etc/fail2ban/jail.local
sudo systemctl restart fail2ban
mkdir /minecraft
curl -sLSf https://launcher.mojang.com/v1/objects/a16d67e5807f57fc4e550299cf20226194497dc2/server.jar > /minecraft/server.jar

echo "eula=true" > /minecraft/eula.txt
mv /tmp/server.properties /minecraft/server.properties
systemctl restart minecraft.service
systemctl enable minecraft.service`

	bedrockCloudInitWant = `#cloud-config
users:
  - default
  - name: prometheus
    shell: /bin/false
  - name: node_exporter
    shell: /bin/false
  

package_update: true

packages:
  - apt-transport-https
  - ca-certificates
  - curl
  - unzip
  - fail2ban

fs_setup:
  - label: minecraft
    device: /dev/sda
    filesystem: xfs
    overwrite: false

mounts:
  - [/dev/sda, /minecraft]

# Enable ipv4 forwarding, required on CIS hardened machines
write_files:
  - path: /etc/sysctl.d/enabled_ipv4_forwarding.conf
    content: |
      net.ipv4.conf.all.forwarding=1
  - path: /tmp/server.properties
    content: |
       level-seed=stackitminecraftrocks
       view-distance=10
       enable-jmx-monitoring=false
       
  - path: /tmp/prometheus.yml
    content: |
      global:
        scrape_interval: 15s

      scrape_configs:
        - job_name: 'prometheus'
          scrape_interval: 5s
          static_configs:
            - targets: ['localhost:9090']
        - job_name: 'node_exporter'
          scrape_interval: 5s
          static_configs:
            - targets: ['localhost:9100']
  - path: /etc/systemd/system/prometheus.service
    content: |
      [Unit]
      Description=Prometheus
      Wants=network-online.target
      After=network-online.target
      [Service]
      User=prometheus
      Group=prometheus
      Type=simple
      ExecStart=/usr/local/bin/prometheus \
          --config.file /etc/prometheus/prometheus.yml \
          --storage.tsdb.path /var/lib/prometheus/ \
          --web.console.templates=/etc/prometheus/consoles \
          --web.console.libraries=/etc/prometheus/console_libraries
      [Install]
      WantedBy=multi-user.target
  - path: /etc/systemd/system/node_exporter.service
    content: |
      [Unit]
      Description=Node Exporter
      Wants=network-online.target
      After=network-online.target
      [Service]
      User=node_exporter
      Group=node_exporter
      Type=simple
      ExecStart=/usr/local/bin/node_exporter
      [Install]
      WantedBy=multi-user.target
  - path: /etc/systemd/system/minecraft.service
    content: |
      [Unit]
      Description=Minecraft Server
      Documentation=https://www.minecraft.net/en-us/download/server
      [Service]
      WorkingDirectory=/minecraft
      Type=simple
      ExecStart=/bin/sh -c "LD_LIBRARY_PATH=. ./bedrock_server"
      Restart=on-failure
      RestartSec=5
      [Install]
      WantedBy=multi-user.target

runcmd:
  - export PROM_VERSION=2.28.1
  - mkdir /etc/prometheus
  - mkdir /var/lib/prometheus
  - curl -sSL https://github.com/prometheus/prometheus/releases/download/v$PROM_VERSION/prometheus-$PROM_VERSION.linux-amd64.tar.gz | tar -xz
  - cp prometheus-$PROM_VERSION.linux-amd64/prometheus /usr/local/bin/
  - cp prometheus-$PROM_VERSION.linux-amd64/promtool /usr/local/bin/
  - chown prometheus:prometheus /usr/local/bin/prometheus
  - chown prometheus:prometheus /usr/local/bin/promtool
  - cp -r prometheus-$PROM_VERSION.linux-amd64/consoles /etc/prometheus
  - cp -r prometheus-$PROM_VERSION.linux-amd64/console_libraries /etc/prometheus
  - chown -R prometheus:prometheus /var/lib/prometheus
  - chown -R prometheus:prometheus /etc/prometheus/consoles
  - chown -R prometheus:prometheus /etc/prometheus/console_libraries
  - mv /tmp/prometheus.yml /etc/prometheus/prometheus.yml
  - chown prometheus:prometheus /etc/prometheus/prometheus.yml
  - systemctl daemon-reload
  - systemctl start prometheus
  - systemctl enable prometheus

  - export NODE_EXPORTER_VERSION=1.1.2
  - curl -sSL https://github.com/prometheus/node_exporter/releases/download/v$NODE_EXPORTER_VERSION/node_exporter-$NODE_EXPORTER_VERSION.linux-amd64.tar.gz | tar -xz
  - cp node_exporter-$NODE_EXPORTER_VERSION.linux-amd64/node_exporter /usr/local/bin
  - chown node_exporter:node_exporter /usr/local/bin/node_exporter
  - systemctl daemon-reload
  - systemctl start node_exporter
  - systemctl enable node_exporter

  - ufw allow ssh
  - ufw allow 5201
  - ufw allow proto udp to 0.0.0.0/0 port 19132
  - echo [DEFAULT] | sudo tee -a /etc/fail2ban/jail.local
  - echo banaction = ufw | sudo tee -a /etc/fail2ban/jail.local
  - echo [sshd] | sudo tee -a /etc/fail2ban/jail.local
  - echo enabled = true | sudo tee -a /etc/fail2ban/jail.local
  - sudo systemctl restart fail2ban
  - curl -sLSf https://minecraft.azureedge.net/bin-linux/bedrock-server-1.17.2.01.zip > /tmp/bedrock-server.zip
  - unzip -o /tmp/bedrock-server.zip -d /minecraft
  - chmod +x /minecraft/bedrock_server
  - echo "eula=true" > /minecraft/eula.txt
  - mv /tmp/server.properties /minecraft/server.properties
  - systemctl restart minecraft.service
  - systemctl enable minecraft.service`

	javaCloudInitWant = `#cloud-config
users:
  - default
  - name: prometheus
    shell: /bin/false
  - name: node_exporter
    shell: /bin/false
  - name: minecraft_exporter
    shell: /bin/false

package_update: true

packages:
  - apt-transport-https
  - ca-certificates
  - curl
  - openjdk-16-jre-headless
  - fail2ban

fs_setup:
  - label: minecraft
    device: /dev/sda
    filesystem: xfs
    overwrite: false

mounts:
  - [/dev/sda, /minecraft]

# Enable ipv4 forwarding, required on CIS hardened machines
write_files:
  - path: /etc/sysctl.d/enabled_ipv4_forwarding.conf
    content: |
      net.ipv4.conf.all.forwarding=1
  - path: /tmp/server.properties
    content: |
       level-seed=stackitminecraftrocks
       view-distance=10
       enable-jmx-monitoring=false
       broadcast-rcon-to-ops=true
       rcon.port=2
       enable-rcon=true
       rcon.password=test
  - path: /tmp/prometheus.yml
    content: |
      global:
        scrape_interval: 15s

      scrape_configs:
        - job_name: 'prometheus'
          scrape_interval: 5s
          static_configs:
            - targets: ['localhost:9090']
        - job_name: 'node_exporter'
          scrape_interval: 5s
          static_configs:
            - targets: ['localhost:9100']
        - job_name: 'minecraft_exporter'
          scrape_interval: 1m
          static_configs:
            - targets: ['localhost:9150']
  - path: /etc/systemd/system/prometheus.service
    content: |
      [Unit]
      Description=Prometheus
      Wants=network-online.target
      After=network-online.target
      [Service]
      User=prometheus
      Group=prometheus
      Type=simple
      ExecStart=/usr/local/bin/prometheus \
          --config.file /etc/prometheus/prometheus.yml \
          --storage.tsdb.path /var/lib/prometheus/ \
          --web.console.templates=/etc/prometheus/consoles \
          --web.console.libraries=/etc/prometheus/console_libraries
      [Install]
      WantedBy=multi-user.target
  - path: /etc/systemd/system/node_exporter.service
    content: |
      [Unit]
      Description=Node Exporter
      Wants=network-online.target
      After=network-online.target
      [Service]
      User=node_exporter
      Group=node_exporter
      Type=simple
      ExecStart=/usr/local/bin/node_exporter
      [Install]
      WantedBy=multi-user.target
  - path: /etc/systemd/system/minecraft-prometheus-exporter.service
    content: |
      [Unit]
      Description=Minecraft Exporter
      Wants=network-online.target
      After=network-online.target
      [Service]
      User=minecraft_exporter
      Group=minecraft_exporter
      Type=simple
      ExecStart=/usr/local/bin/minecraft-prometheus-exporter \
          --mc.rcon-password=test
      [Install]
      WantedBy=multi-user.target
  - path: /etc/systemd/system/minecraft.service
    content: |
      [Unit]
      Description=Minecraft Server
      Documentation=https://www.minecraft.net/en-us/download/server
      [Service]
      WorkingDirectory=/minecraft
      Type=simple
      ExecStart=/usr/bin/java -Xmx2G -Xms2G -jar server.jar nogui
      
      Restart=on-failure
      RestartSec=5
      [Install]
      WantedBy=multi-user.target

runcmd:
  - export PROM_VERSION=2.28.1
  - mkdir /etc/prometheus
  - mkdir /var/lib/prometheus
  - curl -sSL https://github.com/prometheus/prometheus/releases/download/v$PROM_VERSION/prometheus-$PROM_VERSION.linux-amd64.tar.gz | tar -xz
  - cp prometheus-$PROM_VERSION.linux-amd64/prometheus /usr/local/bin/
  - cp prometheus-$PROM_VERSION.linux-amd64/promtool /usr/local/bin/
  - chown prometheus:prometheus /usr/local/bin/prometheus
  - chown prometheus:prometheus /usr/local/bin/promtool
  - cp -r prometheus-$PROM_VERSION.linux-amd64/consoles /etc/prometheus
  - cp -r prometheus-$PROM_VERSION.linux-amd64/console_libraries /etc/prometheus
  - chown -R prometheus:prometheus /var/lib/prometheus
  - chown -R prometheus:prometheus /etc/prometheus/consoles
  - chown -R prometheus:prometheus /etc/prometheus/console_libraries
  - mv /tmp/prometheus.yml /etc/prometheus/prometheus.yml
  - chown prometheus:prometheus /etc/prometheus/prometheus.yml
  - systemctl daemon-reload
  - systemctl start prometheus
  - systemctl enable prometheus

  - export NODE_EXPORTER_VERSION=1.1.2
  - curl -sSL https://github.com/prometheus/node_exporter/releases/download/v$NODE_EXPORTER_VERSION/node_exporter-$NODE_EXPORTER_VERSION.linux-amd64.tar.gz | tar -xz
  - cp node_exporter-$NODE_EXPORTER_VERSION.linux-amd64/node_exporter /usr/local/bin
  - chown node_exporter:node_exporter /usr/local/bin/node_exporter
  - systemctl daemon-reload
  - systemctl start node_exporter
  - systemctl enable node_exporter

  - export MINECRAFT_EXPORTER_VERSION=0.3.1
  - curl -sSL https://github.com/dirien/minecraft-prometheus-exporter/releases/download/v$MINECRAFT_EXPORTER_VERSION/minecraft-prometheus-exporter_$MINECRAFT_EXPORTER_VERSION.linux-amd64.tar.gz | tar -xz
  - cp minecraft-prometheus-exporter /usr/local/bin
  - chown minecraft_exporter:minecraft_exporter /usr/local/bin/minecraft-prometheus-exporter
  - systemctl start minecraft-prometheus-exporter.service
  - systemctl enable minecraft-prometheus-exporter.service
  - ufw allow ssh
  - ufw allow 5201
  - ufw allow proto udp to 0.0.0.0/0 port 25565
  - echo [DEFAULT] | sudo tee -a /etc/fail2ban/jail.local
  - echo banaction = ufw | sudo tee -a /etc/fail2ban/jail.local
  - echo [sshd] | sudo tee -a /etc/fail2ban/jail.local
  - echo enabled = true | sudo tee -a /etc/fail2ban/jail.local
  - sudo systemctl restart fail2ban
  - curl -sLSf https://launcher.mojang.com/v1/objects/a16d67e5807f57fc4e550299cf20226194497dc2/server.jar > /minecraft/server.jar
  
  - echo "eula=true" > /minecraft/eula.txt
  - mv /tmp/server.properties /minecraft/server.properties
  - systemctl restart minecraft.service
  - systemctl enable minecraft.service`
)

func TestCivoBedrockTemplate(t *testing.T) {
	t.Run("Test Template Bedrock for Civo bash", func(t *testing.T) {
		civo, err := NewTemplateCivo(&bedrock)
		if err != nil {
			t.Fatal(err)
		}
		got, err := civo.GetTemplate()
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, bedrockCivoWant, got)
	})
}

func TestCivoJavaTemplate(t *testing.T) {
	t.Run("Test Template Java for Civo bash", func(t *testing.T) {
		civo, err := NewTemplateCivo(&java)
		if err != nil {
			t.Fatal(err)
		}
		got, err := civo.GetTemplate()
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, javaCivoWant, got)
	})
}

func TestCloudInitBedrockTemplate(t *testing.T) {
	t.Run("Test Template Bedrock for Cloud-Init", func(t *testing.T) {
		civo, err := NewTemplateCloudConfig(&bedrock, "sda")
		if err != nil {
			t.Fatal(err)
		}
		got, err := civo.GetTemplate()
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, bedrockCloudInitWant, got)
	})
}

func TestCloudInitJavaTemplate(t *testing.T) {
	t.Run("Test Template Java for Cloud-Init", func(t *testing.T) {
		civo, err := NewTemplateCloudConfig(&java, "sda")
		if err != nil {
			t.Fatal(err)
		}
		got, err := civo.GetTemplate()
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, javaCloudInitWant, got)
	})
}