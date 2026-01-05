import jax
import jax.numpy as jnp
from loguru import logger


jax.distributed.initialize()

logger.info(f"JAX distributed initialized {jax.devices()}")

# Simple model: y = w*x + b
def model(params, x):
    w, b = params
    return w * x + b

def loss_fn(params, x, y):
    pred = model(params, x)
    return jnp.mean((pred - y) ** 2)

# Initialize parameters
params = [1.0, 0.0]

# Dummy data
x = jnp.array([1.0, 2.0, 3.0])
y = jnp.array([2.0, 4.0, 6.0])

# Training loop
learning_rate = 0.1

logger.info("Starting training")

for step in range(10000):
    loss, grads = jax.value_and_grad(loss_fn)(params, x, y)
    if step % 100 == 0:
        logger.info(f"Step {step}: loss = {loss:.4f}")
    
    params = jax.tree_util.tree_map(lambda p, g: p - learning_rate * g, params, grads)

logger.info("Training complete")