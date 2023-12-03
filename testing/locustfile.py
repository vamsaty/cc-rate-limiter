from random import randint

from locust import HttpUser, task, between


class HelloWorldUser(HttpUser):
    # 0.01 seconds between requests i.e. around 100 requests per second
    wait_time = between(0.01, 0.01)

    @task
    def hello_world(self):
        # the server (limiter/test_server.go) uses the X-User header to identify the user
        self.client.get("/limited", headers={
            "X-User": f"user_#_{randint(1, 2)}"
        })
