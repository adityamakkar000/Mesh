<div align="center">

# *MESH*

**A Multi Host JAX Orchestrator **

<br>

<img src="public/mesh.png" alt="MESH logo" width="256" height="256" />

</div>


## Setup


1. Define a `cluster.yaml` in `~/.config/mesh/cluster.yaml` in the following format
```yaml
<cluster_name_1>:  
  user: <user name on tpu>
  identity_file: <path to ssh identity file>
  hosts:
    - <host_1_ip>
    - 34.75.233.113
    - 35.237.15.216
    - ...


<cluster_name_2>:
  ... 
```

2. Define a `mesh.yaml` in your project directory in the following format
```yaml
commands:
  - <setup commands>
  - curl -LsSf https://astral.sh/uv/install.sh | sh

ignore:
  - <files to ignore> 
  - "*.pyc"
  - "*/__pycache__/*"
  - "./git/*"
  - ".env"
  - "mesh.yaml"
  - "*/prerun/*"

prerun:
  - <steps to run before your command>
  - uv venv --python 3.12 --clear
  - source .venv/bin/activate
  - uv pip install loguru 'jax[tpu]'
```

3. Run `mesh setup <cluster>` to run the `commands` from `mesh.yaml` on the `cluster` on every host defined in the `~/.config/cluster.yaml`

Example output 
```sh
╰$ ./mesh setup testWatch           
==> parsed cluster test3 with 32 hosts
==> parsed cluster testReal with 4 hosts
==> parsed cluster testWatch with 1 hosts
==> available clusters: [test3 testReal testWatch]
==> Setting up cluster 'testWatch' (1 hosts)
==> [35.186.108.225] setup completed
==> Cluster 'testWatch' setup complete
```

4. To run a command `mesh run <cluster> <your command>`. This will copy your current directory over to every host in the cluster, run all pre-run commands and then run your command with logs. If you kill the command in your local machine it will mimic this and kill it across every host. The philosophy of MESH is to make it seem you are running single-controller JAX whereas in reality MESH is orchestrating the calls to all the hosts and managing resources, cleanup, etc. 

Note: Note it passes in the rank of the process so host x will run on each cluster 'RANK=x python main.py --lr 1e-3'
Be sure to initalize JAX with induvial ranks using ENV variables to display proper logs from process 0



