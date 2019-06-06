""" GetState
Usage:
    get_state.py <pem-file> [options]

Options:
    -h, --help          print help message and exit
    --output DIR        set the output directory [default: logs]
"""

from docopt import docopt
import paramiko

def get_logs(ip_addr, pem_file, log_dir):
    pem = paramiko.RSAKey.from_private_key_file(pem_file)
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect(hostname=ip_addr, username="ec2-user", pkey=pem)
    ftp = client.open_sftp()
    logs = ftp.listdir('/home/ec2-user/logs/*.txt')
    for l in logs:
        print(l)
        ftp.get(f'/home/ec2-user/logs/{l}', f"{log_dir}/{l}")
    ftp.close()
    client.close()

if __name__ == '__main__':
    args = docopt(__doc__)

    for ip in open('ips.txt'):
        get_logs(ip.strip(), args['<pem-file>'], args['--output'])
