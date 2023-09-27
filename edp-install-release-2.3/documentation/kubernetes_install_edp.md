## EDP Installation on Kubernetes

### Prerequisites
1. Kubernetes cluster installed with minimum 2 worker nodes with total capacity 16 Cores and 40Gb RAM;
2. Machine with [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed with a cluster-admin access to the Kubernetes cluster;
3. Ingress controller is installed in a cluster, for example [ingress-nginx](https://kubernetes.github.io/ingress-nginx/deploy/);
4. Ingress controller is configured with the disabled HTTP/2 protocol and header size of 32k support;

    - Example of Config Map for Nginx ingress controller:
    ```yaml
    kind: ConfigMap
    apiVersion: v1
    metadata:
      name: nginx-configuration
      namespace: ingress-nginx
      labels:
        app.kubernetes.io/name: ingress-nginx
        app.kubernetes.io/part-of: ingress-nginx
    data:
      client-header-buffer-size: 64k
      large-client-header-buffers: 4 64k
      use-http2: "false"
      ```

5. Load balancer (if any exists in front of ingress controller) is configured with session stickiness, disabled HTTP/2 protocol and header size of 32k support;
6. Cluster nodes and pods should have access to the cluster via external URLs. For instance, you should add in AWS your VPC NAT gateway elastic IP to your cluster external load balancers security group);
7. Keycloak instance is installed. To get accurate information on how to install Keycloak, please refer to the [Keycloak Installation on Kubernetes](kubernetes_install_keycloak.md)) instruction;
8. The "openshift" realm is created in Keycloak;
9. The "keycloak" secret with administrative access username and password exists in the namespace where Keycloak in installed;
10. Helm 3 is installed on installation machine with the help of the following [instruction](https://v3.helm.sh/docs/intro/install/).

### EDP namespace
* Clone or download and extract the latest release version that should be installed to a separate folder; 

* Choose an EDP tenant name, e.g. "demo", and create the <edp-project> namespace with any name (e.g. "demo").
Before starting EDP deployment, EDP namespace <edp-project> in K8s should be created.

* Create admin secret for the Wizard database: 
```bash
kubectl -n <edp-project> create secret generic super-admin-db --from-literal=username=<db_admin_username> --from-literal=password=<db_admin_password>
```

* Deploy database from the following template:
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: postgres
---
apiVersion: v1 #PVC for EDP Install Wizard DB
kind: PersistentVolumeClaim
metadata:
  annotations:
    volume.beta.kubernetes.io/storage-provisioner: kubernetes.io/aws-ebs
  finalizers:
    - kubernetes.io/pvc-protection
  name: edp-install-wizard-db
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 2Gi
  storageClassName: gp2
  volumeMode: Filesystem
---
apiVersion: apps/v1 # EDP Install Wizard DB Deployment
kind: Deployment
metadata:
  generation: 1
  labels:
    app: edp-install-wizard-db
  name: edp-install-wizard-db
spec:
  selector:
    matchLabels:
      app: edp-install-wizard-db
  template:
    metadata:
      labels:
        app: edp-install-wizard-db
    spec:
      containers:
        - env:
            - name: POSTGRES_USER
              valueFrom:
                secretKeyRef:
                  key: username
                  name: super-admin-db
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  key: password
                  name: super-admin-db
            - name: PGDATA
              value: /var/lib/postgresql/data/pgdata
            - name: POD_IP
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: status.podIP
            - name: POSTGRES_DB
              value: edp-install-wizard-db
          image: postgres:9.6
          imagePullPolicy: IfNotPresent
          livenessProbe:
            exec:
              command:
                - sh
                - -c
                - exec pg_isready --host $POD_IP -U postgres -d postgres
            failureThreshold: 5
            initialDelaySeconds: 60
            periodSeconds: 20
            successThreshold: 1
            timeoutSeconds: 5
          name: edp-install-wizard-db
          ports:
            - containerPort: 5432
              name: db
              protocol: TCP
          readinessProbe:
            exec:
              command:
                - sh
                - -c
                - exec pg_isready --host $POD_IP -U postgres -d postgres
            failureThreshold: 3
            initialDelaySeconds: 60
            periodSeconds: 20
            successThreshold: 1
            timeoutSeconds: 3
          resources:
            requests:
              memory: 512Mi
          volumeMounts:
            - mountPath: /var/lib/postgresql/data
              name: edp-install-wizard-db
      serviceAccountName: postgres
      volumes:
        - name: edp-install-wizard-db
          persistentVolumeClaim:
            claimName: edp-install-wizard-db
---
apiVersion: v1 # EDP Install Wizard DB Service
kind: Service
metadata:
  name: edp-install-wizard-db
spec:
  ports:
    - name: db
      port: 5432
      protocol: TCP
      targetPort: 5432
  selector:
    app: edp-install-wizard-db
  type: ClusterIP
```

* Create secret for the EDP tenant database user:
```bash
kubectl -n <edp-project> create secret generic admin-console-db --from-literal=username=<tenant_db_username> --from-literal=password=<tenant_db_password>
```
    
### Install EDP
* Deploy operators in the <edp-project> namespace by following the corresponding instructions in their repositories:
    - [keycloak-operator](https://github.com/epmd-edp/keycloak-operator/tree/release-1.3)
    - [codebase-operator](https://github.com/epmd-edp/codebase-operator/tree/release-2.3)
    - [jenkins-operator](https://github.com/epmd-edp/jenkins-operator/tree/release-2.3)
    - [edp-component-operator](https://github.com/epmd-edp/edp-component-operator/tree/release-0.2)    
    - [cd-pipeline-operator](https://github.com/epmd-edp/cd-pipeline-operator/tree/release-2.3)
    - [nexus-operator](https://github.com/epmd-edp/nexus-operator/tree/release-2.3)
    - [sonar-operator](https://github.com/epmd-edp/sonar-operator/tree/release-2.3)
    - [admin-console-operator](https://github.com/epmd-edp/admin-console-operator/tree/release-2.3)
    - [reconciler](https://github.com/epmd-edp/reconciler/tree/release-2.3)

* Create a config map with additional tools (e.g. Sonar, Nexus, Secrets, any other resources) that are non-mandatory.
* Inspect the list of parameters that can be used in the Helm chart and replaced during the provisioning:
    
    - edpName - this parameter will be replaced with the edp.name value, which is set in EDP-Install chart;
    - dnsWildCard - this parameter will be replaced with the edp.dnsWildCard value, which is set in EDP-Install chart;
    - users - this parameter will be replaced with the edp.superAdmins value, which is set in EDP-Install chart. 
    
_*NOTE*: The users parameter should be used in a cycle because it is presented as the list. Other parameters must be hardcorded in a template._
    
Become familiar with a template sample:
```yaml
apiVersion: v2.edp.epam.com/v1alpha1
kind: Nexus
metadata:
  name: nexus
  namespace: '{{ .Values.edpName }}'
spec:
  edpSpec:
    dnsWildcard: '{{ .Values.dnsWildCard }}'
  keycloakSpec:
    enabled: true
  users:
  {{ range .Values.users }}
  - email: ''
    first_name: ''
    last_name: ''
    roles:
      - nx-admin
    username: {{ . }}
  {{ end }}
  image: 'sonatype/nexus3'
  version: 3.21.2
  volumes:
    - capacity: 5Gi
      name: data
      storage_class: gp2
---
apiVersion: v2.edp.epam.com/v1alpha1
kind: Sonar
metadata:
  name: sonar
  namespace: '{{ .Values.edpName }}'
spec:
  edpSpec:
    dnsWildcard: '{{ .Values.dnsWildCard }}'
  type: Sonar
  image: sonarqube
  version: 7.9-community
  initImage: busybox
  dbImage: 'postgres:9.6'
  volumes:
    - capacity: 1Gi
      name: data
      storage_class: gp2
    - capacity: 1Gi
      name: db
      storage_class: gp2
---
apiVersion: v2.edp.epam.com/v1alpha1
kind: GitServer
metadata:
  name: git-epam
  namespace: '{{ .Values.edpName }}'
spec:
  createCodeReviewPipeline: true
  gitHost: 'git.epam.com'
  gitUser: git
  httpsPort: 443
  nameSshKeySecret: gitlab-sshkey
  sshPort: 22
---
apiVersion: v1
data:
  id_rsa: XXXXXXXXXXXXXXXXXXXXXXXX
  id_rsa.pub: XXXXXXXXXXXXXXXXXXXXXXXX
  username: XXXXXXXXXXXXXXXXXXXXXXX
kind: Secret
metadata:
  name: gitlab-sshkey
  namespace: '{{ .Values.edpName }}'
type: Opaque
---
apiVersion: v2.edp.epam.com/v1alpha1
kind: JenkinsServiceAccount
metadata:
  name: gitlab-sshkey
  namespace: '{{ .Values.edpName }}'
spec:
  credentials: 'gitlab-sshkey'
  type: ssh
```

* Create a file with the template and create a config map with the following command:
`kubectl -n <edp-project> create cm additional-tools --from-file=template=<filename>`

* Apply EDP chart using Helm. 

The deploy-templates/values.yaml file contains EDP Helm chart parameters.

>**WARNING**: Chart has some **hardcoded** parameters, which are already fixed in file and are optional for editing, and some **mandatory** parameters that must be specified by user. 
 
Find below the description of both parameters types.

Hardcoded parameters (optional): 
```
    - edp.version - EDP Image and tag. The released version can be found on [Dockerhub](https://hub.docker.com/r/epamedp/edp-install/tags);
    - edp.additionalToolsTemplate - name of the config map in edp-deploy project with a Helm template that is additionally deployed during the installation (Sonar, Gerrit, Nexus, Secrets, Any other resources). **You created it in the previous point.**
    - edp.devDeploy: Used for develomplent deploy using CI for production installation should be false;
    - jenkins.version - EDP image and tag. The released version can be found on [Dockerhub](https://hub.docker.com/r/epamedp/edp-jenkins/tags);
    - jenkins.volumeCapacity - size of persistent volume for Jenkins data, it is recommended to use not less then 10 GB;
    - jenkins.stagesVersion - version of EDP-Stages library for Jenkins. The released version can be found on [GitHub](https://github.com/epmd-edp/edp-library-stages/releases);
    - jenkins.pipelinesVersion - version of EDP-Pipeline library for Jenkins. The released version can be found on [GitHub](https://github.com/epmd-edp/edp-library-pipelines/releases);
    - adminConsole.version - EDP image and tag. The released version can be found on [Dockerhub](https://hub.docker.com/r/epamedp/edp-admin-console/tags);
    - perf.* - Integration with PERF is in progress. Should be false so far;
```
 
Mandatory parameters:
 ```
    - edp.name - previously defined name of your EDP project <edp-project>;
    - edp.superAdmins - administrators of your tenant separated by escaped comma (\,);
    - edp.dnsWildCard - DNS wildcard for routing in your K8S cluster;
    - edp.storageClass - storage class that will be used for persistent volumes provisioning;
    - edp.platform - openshift or kubernetes
    - edp.keycloakNamespace: namespace where Keycloak is installed;
    - edp.keycloakUrl: FQDN Keycloak URL.
    - edp.webConsole - Kubernetes web console URL (e.g. https://master.example.com:8443); 
 ```  
 
 * Edit deploy-templates/values.yaml file with your own parameters;
 * Run Helm chart installation;

Find below the sample of launching a Helm template for EDP installation:
```bash
helm install edp-install --namespace <edp-project> deploy-templates
```
 * Add the EDP Component CR to the namespace (<edp-project>) that was created after the installation for Docker Registry with its specified URL.
```yaml
apiVersion: v1.edp.epam.com/v1alpha1
kind: EDPComponent
metadata:
  name: arbitraryName
 spec:
  icon: bas64 encoded image
  type: docker-registry #this value should be exactly the same
  url: url for docker registry
```
>_**NOTE**: The full installation with integration between tools can take at least 10 minutes._