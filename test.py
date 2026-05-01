from mimetypes import init
import jax
import os
from loguru import logger


def init_distributed_jax():
    """Initializes JAX distributed environment."""
    RANK = os.environ.get("RANK", None)
    if not RANK and jax.process_index() == 0:
        logger.warning("JAX distributed got no RANK env variable")
    jax.distributed.initialize(process_id=int(RANK) if RANK else None)

    if jax.process_index() == 0:
        process_count = jax.process_count()
        local_devices = len(jax.local_devices())
        logger.info(f"JAX distributed initialized with {process_count} processes with {local_devices} per host.")
        all_devices = jax.devices()
        logger.info("Devices:")
        for dev in all_devices:
            logger.info(f"\tDevice ID: {dev.id}, Platform: {dev.platform}, Kind: {dev.device_kind}")

init_distributed_jax()