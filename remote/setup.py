"""Setup
Usage:
    setup.py <network_file> <pem_file> [options]

Options:
    -h, --help      Print help message and exit
"""

from docopt import docopt
from os import path
import json
import paramiko
import time

FVP_PATH = "/home/ec2-user/go-workplace/src/github.com/kpister/fvp"

if __name__ == '__main__':
    args = docopt(__doc__)
    
    if not path.exists(args['<network_file>']):
        raise Exception("Network file does not exist")

    if not path.exists(args['<pem_file>']):
        raise Exception(".pem file does not exist")
    
    pem = paramiko.RSAKey.from_private_key_file(args['<pem_file>'])

    network_config = [line.strip() for line in open(args['<network_file>']).read().split("\n") if len(line) > 1]
    network_size = len(network_config)

    if network_size < 2:
        raise Exception("Network size cannot be less than 2")

    ips = {}
    for line in network_config:
        host, cfg = line.split('~')
        ip, port = host.split(':')
        if ip not in ips:
            ips[ip] = []
        ips[ip].append((f'{ip}:{port}', cfg))

    for ip in ips.keys():
        # connect to server
        print(f'Connecting to {ip}')
        client = paramiko.SSHClient()
        client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        client.connect(hostname=ip, username="ec2-user", pkey=pem)

        # clear cfgs folder
        cmd = f"rm -f /home/ec2-user/cfgs/*"
        print(f'Executing: {cmd}')
        client.exec_command(cmd)

        for host, cfg in ips[ip]:
            # copy in needed cfgs
            cmd = f"echo '{cfg}' > /home/ec2-user/cfgs/{host}.json"
            print(f'Executing: {cmd}')
            client.exec_command(cmd)

        cmd = "/home/ec2-user/start.sh\n"

        transport = client.get_transport()
        channel = transport.open_session()
        pty = channel.get_pty()
        shell = client.invoke_shell()
        shell.send(cmd)
        time.sleep(5)

        shell.close()
        channel.close()
        client.close()
