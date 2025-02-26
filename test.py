import docker
import logging
import os
import time
from flask import Flask, request

logger = logging.getLogger(__name__)

IMAGE_TO_WATCH = 'shlomi1126/rail-go'
CONTAINER_NAME = 'my_container'
WEBHOOK_PORT = 5000

app = Flask(__name__)

def pull_image(tag='latest'):
    """
    Pull the specified image from Docker Hub with retry logic.
    """
    client = docker.from_env()
    username = 'shlomi126'
    password = 'ydr5yvz@hnq_RGT*rqa'
    client.login(username=username, password=password)
    image_name = f"{IMAGE_TO_WATCH}:{tag}"
    logger.info(f"Pulling image: {image_name}")

    max_attempts = 5
    delay = 15  # seconds

    for attempt in range(1, max_attempts + 1):
        try:
            client.images.pull(image_name)
            logger.info(f"Image pulled successfully: {image_name}")
            run_container(image_name)
            break  # Exit the loop if successful
        except docker.errors.APIError as e:
            logger.error(f"Attempt {attempt} failed to pull image: {e}")
            if attempt < max_attempts:
                logger.info(f"Retrying in {delay} seconds...")
                time.sleep(delay)
            else:
                logger.error("Exceeded maximum retry attempts. Exiting.")

def run_container(image_name):
    """
    Stop and remove the previous container, remove the previous image, then run a new container from the specified image.
    """
    client = docker.from_env()

    # Stop and remove the previous container
    try:
        container = client.containers.get(CONTAINER_NAME)
        container.stop()
        container.remove()
        logger.info(f"Stopped and removed container: {CONTAINER_NAME}")
    except docker.errors.NotFound:
        logger.info(f"No existing container found with name: {CONTAINER_NAME}")

    # Remove the previous image
    try:
        client.images.remove(image=f"{IMAGE_TO_WATCH}:latest", force=True)
        logger.info(f"Removed image: {IMAGE_TO_WATCH}:latest")
    except docker.errors.ImageNotFound:
        logger.info(f"No existing image found with name: {IMAGE_TO_WATCH}:latest")

    # Run a new container from the specified image
    logger.info(f"Running container from image: {image_name}")
    try:
        client.containers.run(
            image_name,
            detach=True,
            name=CONTAINER_NAME,
            ports={'5000/tcp': 5000}  # Adjust ports as needed
        )
        logger.info(f"Container started from image: {image_name}")
    except docker.errors.APIError as e:
        logger.error(f"Failed to run container: {e}")

@app.route('/', methods=['POST'])
def webhook():
    data = request.get_json()
    # Optionally extract the tag from the payload
    tag = 'latest'  # Or extract from data if available
    pull_image(tag)
    return 'OK', 200

if __name__ == '__main__':
    logging.basicConfig(level=logging.INFO)
    app.run(host='0.0.0.0', port=WEBHOOK_PORT)

