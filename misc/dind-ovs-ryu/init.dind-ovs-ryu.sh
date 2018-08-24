#!/bin/bash
OVSBR0_IP=${OVSBR0_IP:-192.168.42.254}
OVSBR0_NETMASK=${OVSBR0_NETMASK:-24}
OVSBR0_PROTO=${OVSBR0_PROTO:-OpenFlow13}

/etc/init.d/openvswitch-switch start

(ovs-vsctl list-br | grep ovsbr0 ) || \
    (
	ovs-vsctl add-br ovsbr0
	ovs-vsctl set bridge ovsbr0 protocols=${OVSBR0_PROTO}
	ovs-vsctl set-controller ovsbr0 tcp:127.0.0.1
    )

ip link set ovsbr0 up
ip addr add ${OVSBR0_IP}/${OVSBR0_NETMASK} dev ovsbr0
echo "Assigned ${OVSBR0_IP} to ovsbr0"

[[ $1 ]] && exec "$@"
exec bash --login
