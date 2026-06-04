# Projeto Korp — Desafio Técnico

Infraestrutura completa com serviço HTTP em Go, NGINX como proxy reverso,
monitoramento com Prometheus + Grafana e automação via Ansible.

---

## Estrutura do Projeto

```
projeto-korp/
├── app/
│   ├── main.go              # Serviço HTTP em Go
│   ├── go.mod               # Módulo Go
│   └── Dockerfile           # Imagem multi-stage (build + runtime scratch)
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
│           └── http-server-projeto-korp-dashboard.json  # Dashboard pronto
├── docker-compose.yml       # Orquestra todos os 4 containers
└── ansible/
    ├── inventory.ini
    ├── playbook.yml          # Ponto de entrada único
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
        ├── docker/tasks/main.yml    # Instala Docker + cria rede
        ├── app/tasks/main.yml       # Build da imagem + docker compose up
        └── monitoring/tasks/main.yml # Valida Prometheus e Grafana
```

---

## Pré-requisitos

| Ferramenta    | Versão mínima |
|---------------|---------------|
| Ansible       | 2.14+         |
| Python        | 3.8+          |
| Coleção community.docker | `ansible-galaxy collection install community.docker` |

---

## Provisionamento automático (Ansible)

```bash
# 1. Instalar coleção necessária
ansible-galaxy collection install community.docker

# 2. Executar o playbook (único comando)
cd ansible
ansible-playbook -i inventory.ini playbook.yml
```

O playbook irá:
1. Instalar Docker Engine + Compose Plugin (Debian ou RHEL)
2. Criar a rede bridge `korp-net`
3. Fazer o build da imagem `http-server-projeto-korp:latest`
4. Copiar arquivos de configuração de monitoramento
5. Subir todos os containers via `docker compose`
6. Aguardar o serviço ficar saudável
7. Validar Prometheus e Grafana
8. Realizar requisição HTTP e exibir resposta no console

---

## Execução manual

```bash
# Build da imagem
docker build -t http-server-projeto-korp:latest ./app

# Criar rede (caso não exista)
docker network create --driver bridge korp-net

# Subir toda a stack
docker compose up -d

# Testar
curl http://localhost:80/projeto-korp
```

### Resposta esperada

```json
{
  "nome": "Projeto Korp",
  "horario": "2024-01-15T14:32:00Z"
}
```

---

## Endpoints do serviço (porta 8080, interno)

| Endpoint         | Descrição                        |
|------------------|----------------------------------|
| `GET /projeto-korp` | Retorna nome + horário UTC    |
| `GET /health`    | Health check (`{"status":"ok"}`) |
| `GET /metrics`   | Métricas no formato Prometheus   |

---

## Métricas implementadas

| Métrica               | Tipo    | Descrição                              |
|-----------------------|---------|----------------------------------------|
| `service_up`          | Gauge   | 1 = serviço UP, 0 = DOWN              |
| `http_requests_total` | Counter | Total de requisições por method/path/status |

---

## Acesso aos serviços

| Serviço    | URL                    | Credenciais   |
|------------|------------------------|---------------|
| Aplicação  | http://localhost:80    | —             |
| Prometheus | http://localhost:9090  | —             |
| Grafana    | http://localhost:3000  | admin / admin |

---

## Decisões técnicas

### Imagem Docker em `scratch`
A aplicação Go é compilada com `CGO_ENABLED=0` e a imagem final usa `scratch`
(completamente vazia), resultando em uma imagem de ~10 MB sem shell, sem SO base,
minimizando superfície de ataque.

### Rede bridge isolada
O container `http-server-projeto-korp` **não expõe portas ao host**. Todo o
tráfego externo passa pelo NGINX, que atua como único ponto de entrada.

### Métricas com Prometheus client_golang
O uso da biblioteca oficial evita implementar o formato text/plain do Prometheus
manualmente, garante corretude e é o padrão de mercado para Go.

### Dashboard provisionado como código
O arquivo JSON do dashboard é carregado automaticamente pelo Grafana via
`provisioning/dashboards`, eliminando a necessidade de importação manual e
permitindo versionamento em Git (GitOps).

### Ansible com roles
A separação em roles (`docker`, `app`, `monitoring`) facilita reutilização,
manutenção e testes independentes de cada responsabilidade.
