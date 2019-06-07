import argparse
import random 
import json

"""
Usage:
    gen_qs_cfg.py <network_size> <ip_file> <output_qs_cfg> [options]

Options:
    -m, --max_n INT    Set the max number of servers per instance [default: 5]
    -f, --failure_type      The type of failure
    -e, --evail_prob        Probability to be evil
    -mqz, --max_n_qs        Max number of quorum slices
    -mqsz, --max_qs_size    Max size of quorum slice
    -bt, --broadcast_time   Timeout to send
    -h, --help              Print help message and exit
"""

def main():
    parser = argparse.ArgumentParser(description="gen_qs_cfg.py")
    parser.add_argument("network_size", metavar="network_size", type=int, help="Number of nodes in the network")
    parser.add_argument("ip_file", metavar="ip_file", type=argparse.FileType("r"), help="A list of instances' IP addresses")
    parser.add_argument("output_qs_cfg", metavar="qs_cfg", type=argparse.FileType("w"), help="Output config path")
    parser.add_argument("-m","--max_n", metavar="max_n", nargs="?", default=5, type=int, help="Max number of servers per instance")
    parser.add_argument("-f", "--failure_type", metavar="failure_type", nargs="?", default="random", type=str, help="The type of failure")
    parser.add_argument("-e", "--evil_prob", metavar="evil_prob", nargs="?", default=0, type=int, help="Probability to be evil")
    parser.add_argument("-mqn", "--max_n_qs", metavar="max_n_qs", nargs="?", default=1, type=int, help="Max number of quorum slices")
    parser.add_argument("-mqsz", "--max_qs_size", metavar="max_qs_size", nargs="?", default=3, type=int, help="Max size of quorum slice")
    parser.add_argument("-bt", "--broadcast_time", metavar="broadcast_time", nargs="?", default=10, type=int, help="Timeout to send")
    args = parser.parse_args()

    if args.network_size < 1:
        raise Exception("Network size cannot be less than 1")

    if args.max_n < 1:
        raise Exception("Max size cannot be less than 1")

    if not args.failure_type in ["STOP", "RAND_PUT", "RAND_ALL", "random"]:
        raise Exception("Failure type should be STOP, RAND_PUT, or RAND_ALL")

    if args.evil_prob < 0 or args.evil_prob > 100:
        raise Exception("Evil probability should be within [0, 100]")
    args.evil_prob /= 100.0

    ips = []
    for line in args.ip_file:
        ips.append(line.strip())

    addrs = [] # store a list of all the ip address with ports, for config files
    for i in range(args.network_size):
        ip = ips[i//args.max_n]
        addrs.append(f'{ip}:{8000 + (i % args.max_n)}')

    for node_i in range(args.network_size):
        is_evil = random.random() < args.evil_prob
        if is_evil:
            failure_type = args.failure_type
        else:
            failure_type = "NONE"

        node_qs_cfg = {
            "QsSlices": [],
            "IsEvil": is_evil,
            "Strategy": failure_type,
            "ID": addrs[node_i],
            "BroadcastTimeout": args.broadcast_time,
        }
        n_qs = random.randint(1, args.max_n_qs)

        for _ in range(n_qs):
            qs_sz = random.randint(1, args.max_qs_size)
            _addrs = addrs[:node_i] + addrs[node_i+1:]
            random.shuffle(_addrs)
            qs = _addrs[:qs_sz] + [addrs[node_i]]
            node_qs_cfg["QsSlices"].append(qs)

        args.output_qs_cfg.write(addrs[node_i] + "~" + json.dumps(node_qs_cfg) + "\n")

if __name__ == "__main__":
    main()
