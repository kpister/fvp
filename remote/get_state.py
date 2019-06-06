""" GetState
Usage:
    get_state.py <pem-file> <ip-file> [options]

Options:
    -h, --help          print help message and exit
    --output DIR        set the output directory [default: logs]
"""

from docopt import docopt
import paramiko
import os

def get_logs(ip_addr, pem_file, log_dir):
    pem = paramiko.RSAKey.from_private_key_file(pem_file)
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect(hostname=ip_addr, username="ec2-user", pkey=pem)
    ftp = client.open_sftp()
    logs = sorted(ftp.listdir('/home/ec2-user/logs/'))
    for l in logs:
        if l.endswith('.txt'):
            print(l)
            client.exec_command(f'cat /home/ec2-user/logs/{l} > /home/ec2-user/logs/tmp')
            ftp.get(f'/home/ec2-user/logs/tmp', f"{log_dir}/{l}")
            client.exec_command('rm /home/ec2-user/logs/tmp')
    ftp.close()
    client.close()

if __name__ == '__main__':
    args = docopt(__doc__)

    for ip in open(args['<ip-file>']):
        os.system(f"scp -i {args['<pem-file>']} ec2-user@{ip.strip()}:~/logs/*.txt {args['--output']}")
        #get_logs(ip.strip(), args['<pem-file>'], args['--output'])
