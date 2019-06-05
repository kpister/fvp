""" Setup
Usage:
    install.py <pem-file> <ip-addr>... [options]

Options:
    -h, --help          print help message and exit
    -s, --server        only install server setup
    -c, --client        only install client setup
"""

from docopt import docopt
import paramiko
import time

def install(ip_addr, pem_file):
    print(f'Installing {ip_addr}')
    pem = paramiko.RSAKey.from_private_key_file(pem_file)
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect(hostname=ip_addr, username="ec2-user", pkey=pem)
    ftp = client.open_sftp()
    ftp.put('aux/install.sh', '/home/ec2-user/install.sh')
    ftp.put('aux/start.sh', '/home/ec2-user/start.sh')
    ftp.close()

    transport = client.get_transport()
    channel = transport.open_session()
    pty = channel.get_pty()
    shell = client.invoke_shell()
    shell.send("chmod +x /home/ec2-user/install.sh\n")
    shell.send("chmod +x /home/ec2-user/start.sh\n")
    shell.send("nohup /home/ec2-user/install.sh &\n")
    time.sleep(2)
    shell.close()
    channel.close()
    client.close()


def main(args):
    args = docopt(__doc__)

    for ip in args['<ip-addr>']:
        install(ip, args['<pem-file>'])
    print("All servers are being setup, please allow a minute to finish install")

if __name__ == '__main__':
    main(docopt(__doc__))
