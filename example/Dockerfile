FROM    ubuntu:22.04

COPY ./ms-user/charts ./ms-user/charts
COPY ./ms-policy/charts ./ms-policy/charts
COPY ./ms-api/charts ./ms-api/charts
COPY ./Makefile ./Makefile

RUN     apt-get -y update
RUN     apt-get -y install \
        ca-certificates \
        curl \
        gnupg \
        make \
        lsb-release

RUN     mkdir -p /etc/apt/keyrings
RUN     curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg

RUN     echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
        $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

RUN     apt-get -y update
RUN     apt-get -y install docker-ce docker-ce-cli containerd.io docker-compose-plugin

RUN     echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.28/deb/ /" | tee /etc/apt/sources.list.d/kubernetes.list
RUN     curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.28/deb/Release.key | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

RUN     apt-get update
RUN     apt-get install -y kubectl

RUN     apt-get install conntrack

RUN     curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && chmod +x minikube

RUN     mkdir -p /usr/local/bin/
RUN     install minikube /usr/local/bin/

CMD     tail -F /dev/null