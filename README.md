<div align="center">

# *MESH*
<img src="public/mesh.png" alt="MESH logo" width="256" height="256" />

<br>


**A Multi-Host JAX Orchestrator**

</div>


## Install

Currently only supported on MacOS right now. 

1. Install using `curl -fsSL https://raw.githubusercontent.com/adityamakkar000/mesh/main/scripts/install.sh | bash`
2. Build from source `git clone https://github.com/adityamakkar000/Mesh.git && cd Mesh && bash scripts/install.sh`

Afterwards source your `.zshrc`

## Setup

1. Define a `cluster.yaml` in `~/.config/mesh/cluster.yaml` in the following format:
```yaml
<cluster_name_1>:  
  user: <user name on TPU>
  identity_file: <path to SSH identity file>
  hosts:
    - <host_1_ip>
    - 34.75.233.113
    - 35.237.15.216
    - ...

<cluster_name_2>:
  ...
```

2. Define a `mesh.yaml` in your project directory in the following format:
```yaml
commands:
  - <setup commands>
  - curl -LsSf https://astral.sh/uv/install.sh | sh

ignore:
  - <files to ignore>
  - "*.pyc"
  - "*/__pycache__/*"
  - ".git/*"
  - ".env"
  - "mesh.yaml"
  - "*/prerun/*"

prerun:
  - <steps to run before your command>
  - uv venv --python 3.12 --clear
  - source .venv/bin/activate
  - uv pip install loguru 'jax[tpu]'
```

3. Run `mesh setup <cluster>` to execute the `commands` from `mesh.yaml` on the specified `cluster`, on every host defined in `~/.config/mesh/cluster.yaml`.

Example output:
```sh
╰$ ./mesh setup testWatch
==> parsed cluster test3 with 32 hosts
==> parsed cluster testReal with 4 hosts
==> parsed cluster testWatch with 1 host
==> available clusters: [test3 testReal testWatch]
==> Setting up cluster 'testWatch' (1 host)
==> [35.186.108.225] setup completed
==> Cluster 'testWatch' setup complete
```

4. To run a command, use `mesh run <cluster> <your command>`. This will copy your current directory to every host in the cluster, run all pre-run commands, and then execute your command with logs. If you kill the command on your local machine, MESH will mirror this behavior and terminate it across every host.

The philosophy of MESH is to make it feel like you are running single-controller JAX, while in reality MESH orchestrates calls across all hosts and manages resources, cleanup, etc.

**Note:** Calling `mesh run <cluster> <your command>` passes the rank of the process to each host. For example, host `x` will run:
```
RANK=x python main.py --lr 1e-3 ...
```

Be sure to initialize JAX with individual ranks using environment variables to ensure proper logging from process 0.
