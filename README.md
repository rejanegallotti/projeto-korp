# Projeto Korp — Desafio Técnico DevOps

Infraestrutura completa com serviço HTTP em Go, NGINX como proxy reverso,
monitoramento com Prometheus + Grafana e automação via Ansible.

---

## Estrutura do Projeto

```
projeto-korp/
├── app/
│   ├── main.go              # Serviço HTTP em Go
│   ├── go.mod               # Módulo Go
│   ├── go.sum               # Checksums das dependências
│   └── Dockerfile           # Imagem multi-stage (build + runtime distroless)
├── nginx/
│   └── conf.d/
│       └── http-server-projeto-korp.conf   # Configuração do proxy reverso
├── prometheus/
│   └── prometheus.yml       # Coleta de métricas do serviço
├── grafana/
│   └── provisioning/
│       ├── datasources/
│       │   └── datasources.yml              # Prometheus como datasource padrão
│       └── dashboards/
│           ├── dashboards.yml               # Provider de dashboards
│           └── http-server-projeto-korp-dashboard.json  # Dashboard provisionado
├── docker-compose.yml       # Orquestra todos os 4 containers
└── ansible/
    ├── inventory.ini        # Inventário local
    ├── playbook.yml         # Ponto de entrada único
    ├── files/
    │   ├── prometheus/
    │   │   └── prometheus.yml
    │   └── grafana/
    │       └── provisioning/
    │           ├── datasources/datasources.yml
    │           └── dashboards/
    │               ├── dashboards.yml
    │               └── http-server-projeto-korp-dashboard.json
    └── roles/
        ├── docker/tasks/main.yml     # Instala Docker Engine + Compose Plugin
        ├── app/tasks/main.yml        # Build da imagem + docker compose up
        └── monitoring/tasks/main.yml # Valida Prometheus, Grafana e dashboard
```

---

## Pré-requisitos

| Ferramenta | Versão mínima | Instalação |
|---|---|---|
| Ansible | 2.14+ | `sudo apt install ansible -y` |
| Python | 3.8+ | Incluído no Ubuntu |
| community.docker | qualquer | `ansible-galaxy collection install community.docker` |
| Docker | 24.0+ | Instalado automaticamente pelo playbook |

> **Nota:** O Docker **não precisa estar instalado** previamente — o playbook Ansible instala e configura o Docker automaticamente.

---

## Provisionamento automático (Ansible)

### 1. Instalar a coleção necessária

```bash
ansible-galaxy collection install community.docker
```

### 2. Executar o playbook

```bash
cd ansible
ansible-playbook -i inventory.ini playbook.yml --ask-become-pass
```

> **Sobre o `--ask-become-pass`:** O playbook precisa de privilégios de administrador (sudo) para instalar o Docker e configurar o sistema. Ao rodar o comando, será solicitada a senha sudo do seu usuário. Digite a senha do seu ambiente e pressione Enter — nada será exibido na tela durante a digitação, isso é normal.

O playbook executa automaticamente:
1. Verifica o sistema operacional (suporta Debian/Ubuntu e RHEL/CentOS)
2. Instala Docker Engine + Compose Plugin via repositório oficial
3. Garante que o serviço Docker está habilitado e em execução
4. Cria os diretórios de configuração com permissões corretas
5. Copia arquivos de configuração do Prometheus e Grafana
6. Faz o build da imagem `http-server-projeto-korp:latest`
7. Sobe todos os containers via `docker compose`
8. Valida que a aplicação está respondendo
9. Valida que Prometheus está coletando métricas
10. Valida que Grafana está online com datasource e dashboard configurados
11. Realiza requisição HTTP e exibe a resposta no console

### Exemplo de saída esperada

```
TASK [monitoring : Exibir status do monitoramento]
ok: [localhost] => {
    "msg": [
        "Prometheus:  ONLINE ✓",
        "Grafana:     ONLINE ✓",
        "Datasource:  CONFIGURADO ✓",
        "Dashboard:   PROVISIONADO ✓",
        "service_up = 1"
    ]
}

TASK [Exibir resposta do serviço no console]
ok: [localhost] => {
    "msg": "Resposta de http://localhost:80/projeto-korp
    {
        \"horario\": \"2026-06-05T23:28:15Z\",
        \"nome\": \"Projeto Korp\"
    }"
}

PLAY RECAP
localhost: ok=33  changed=6  failed=0
```

---

## Execução manual (sem Ansible)

```bash
# Build da imagem
docker build -t http-server-projeto-korp:latest ./app

# Subir toda a stack
docker compose up -d

# Testar
curl http://localhost:80/projeto-korp
```

---

## Endpoints do serviço

| Endpoint | Descrição |
|---|---|
| `GET /projeto-korp` | Retorna `{"nome":"Projeto Korp","horario":"<UTC>"}` |
| `GET /health` | Health check — retorna `{"status":"ok"}` |
| `GET /metrics` | Métricas no formato Prometheus |

### Resposta esperada

```json
{"nome":"Projeto Korp","horario":"2026-06-05T23:28:15Z"}
```

---

## Acesso aos serviços

| Serviço | URL | Credenciais |
|---|---|---|
| Aplicação (via NGINX) | http://localhost:80 | — |
| Prometheus | http://localhost:9090 | — |
| Grafana | http://localhost:3000 | admin / admin |

---

## Métricas implementadas

| Métrica | Tipo | Descrição |
|---|---|---|
| `service_up` | Gauge | 1 = serviço UP, 0 = DOWN |
| `http_requests_total` | Counter | Total de requisições por method, path e status HTTP |

---

## Decisões técnicas

### Imagem Docker distroless com usuário não-root
A imagem final usa `gcr.io/distroless/static:nonroot` — sem shell, sem
gerenciador de pacotes, sem ferramentas. O processo roda com UID 65532
(usuário `nonroot`), sem privilégios de root. Princípio do menor privilégio.

### Aplicação sem porta exposta ao host
O container `http-server-projeto-korp` não expõe portas diretamente ao host.
Todo o tráfego externo passa obrigatoriamente pelo NGINX — único ponto de
entrada controlado.

### Rede bridge isolada (korp-net)
Rede Docker customizada com DNS automático entre containers. O NGINX referencia
a aplicação por nome (`http-server-projeto-korp:8080`), não por IP.

### Dashboard Grafana provisionado como código (GitOps)
O dashboard é carregado automaticamente via arquivos de provisioning —
`datasources.yml`, `dashboards.yml` e o JSON do dashboard. Nenhuma configuração
manual necessária. A configuração está versionada no Git.

### Ansible com roles separadas por responsabilidade
Roles `docker`, `app` e `monitoring` com responsabilidade única. O role
`monitoring` valida não só que os serviços estão rodando, mas que o datasource
e o dashboard foram provisionados corretamente.

### Graceful shutdown e timeouts HTTP
O servidor captura SIGTERM/SIGINT e encerra conexões de forma organizada com
timeout de 30 segundos. Timeouts de ReadTimeout (5s), WriteTimeout (10s) e
IdleTimeout (120s) protegem contra ataques Slowloris.