[English](README.md) | Español

# exposedwatch

[![CI](https://github.com/Byroncl/exposedwatch/actions/workflows/ci.yml/badge.svg)](https://github.com/Byroncl/exposedwatch/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Byroncl/exposedwatch)](https://github.com/Byroncl/exposedwatch/releases/latest)
[![License](https://img.shields.io/github/license/Byroncl/exposedwatch)](LICENSE)

Muestra en tu terminal, en tiempo real, qué expone tu servidor Linux a la red y avisa cuando hay algo riesgoso abierto, como una base de datos enlazada a `0.0.0.0`.

Lee los puertos en escucha del host (desde `/proc/net`) y los puertos publicados de los contenedores Docker.

## Inicio rápido

En Ubuntu/Debian (`amd64` o `arm64`):

```sh
curl -fLO "https://github.com/Byroncl/exposedwatch/releases/latest/download/exposedwatch_linux_$(dpkg --print-architecture).deb"
sudo apt install "./exposedwatch_linux_$(dpkg --print-architecture).deb"
sudo exposedwatch
```

`sudo exposedwatch` abre una vista a pantalla completa que se actualiza sola (como `htop`). Pulsa `?` para ver la ayuda y `q` para salir.

| Comando | Qué hace |
| ------- | -------- |
| `exposedwatch` | Vista interactiva a pantalla completa (igual que `exposedwatch tui`) |
| `exposedwatch scan` | Informe de una sola ejecución para scripts y cron (tabla o JSON) |
| `exposedwatch version` | Muestra la versión (también `--version`) |
| `exposedwatch help` | Muestra la ayuda (también `-h`, `--help`) |

Las opciones indicadas al inicio, como `exposedwatch --interval 5s`, se pasan a la vista interactiva.

## Qué detecta

Las reglas se aplican por puerto. Una regla se activa cuando el puerto está expuesto al menos tanto como la exposición mínima indicada.

| Servicio | Puerto(s) | Exposición mínima | Severidad |
| -------- | --------- | ----------------- | --------- |
| SSH | 22 | `all-interfaces` | medium |
| Telnet | 23 | `all-interfaces` | critical |
| Base de datos o caché: MySQL, PostgreSQL, Redis, MongoDB, Elasticsearch, Memcached | 3306, 5432, 6379, 27017, 9200, 11211 | `private` | high |
| API de Docker (sin autenticación) | 2375 | `all-interfaces` | critical |
| RDP | 3389 | `all-interfaces` | high |
| SMB / NetBIOS | 445, 139 | `all-interfaces` | high |
| Cualquier otro puerto | cualquiera | `all-interfaces` | info |

Niveles de exposición, de menor a mayor alcance: `loopback`, `private`, `all-interfaces`, `public` (enlazado a una dirección IP pública concreta). Cada hallazgo incluye una recomendación.

Severidades: `info`, `low`, `medium`, `high`, `critical`.

## Vista interactiva

```sh
sudo exposedwatch                      # se actualiza cada 2 s
sudo exposedwatch --interval 5s        # intervalo personalizado (mínimo 1s)
sudo exposedwatch --no-docker          # omite el descubrimiento de Docker
```

Opciones: `--proc-root`, `--docker-socket`, `--no-docker`, `--interval`. Requiere una terminal interactiva; de lo contrario termina con código 2 (usa `scan`). Si una fuente falla (por ejemplo, Docker no disponible), se muestra como una línea de advertencia y la vista continúa.

| Tecla | Acción |
| ----- | ------ |
| `up` / `k`, `down` / `j` | Mover la selección |
| `g` / `G` | Ir al inicio / al final |
| `r` | Actualizar ahora |
| `s` | Cambiar el orden: severidad, puerto, proceso |
| `f` | Cambiar el filtro de severidad mínima |
| `/` | Filtro de texto libre (`enter` aplica, `esc` cancela) |
| `?` | Mostrar u ocultar la ayuda |
| `q` / `ctrl+c` | Salir |

## Scripts, cron y CI: `scan`

```sh
sudo exposedwatch scan                       # salida en tabla
sudo exposedwatch scan --format json         # salida legible por máquinas
sudo exposedwatch scan --min-severity medium # muestra y evalúa desde medium en adelante
sudo exposedwatch scan --no-docker           # omite el descubrimiento de Docker
```

Ejemplo de salida:

```text
$ exposedwatch scan --no-docker --min-severity high
SEVERITY  PORT  PROTOCOL  BIND ADDRESS  EXPOSURE        SOURCE  PROCESS/CONTAINER  TITLE
high      5432  tcp       0.0.0.0       all-interfaces  host    -                  Database or cache reachable beyond loopback
high      5432  tcp       ::            all-interfaces  host    -                  Database or cache reachable beyond loopback

Recommendations
  - [database-exposed] Bind the datastore to 127.0.0.1 or an internal network and require authentication.

54 port(s) scanned, 2 finding(s).
```

Códigos de salida (comunes a todos los comandos):

| Código | Significado |
| ------ | ----------- |
| 0 | Correcto, sin hallazgos en `--min-severity` o superior |
| 1 | Error de ejecución |
| 2 | Error de uso (opción o comando inválido) |
| 3 | Se encontraron hallazgos en `--min-severity` o superior |

Ejemplo para cron o CI: hacer fallar el trabajo si hay algo de severidad medium o mayor expuesto.

```sh
exposedwatch scan --min-severity medium --format json > exposure.json || echo "exposure check exited with $?"
```

## Escaneo externo (nmap)

`--external` también escanea esta máquina desde el lado de la red con [nmap](https://nmap.org) e informa lo que encuentra con la fuente `external`, combinado con los resultados del host y de Docker y evaluado con las mismas reglas de riesgo. Está desactivado por defecto.

```sh
sudo apt install nmap
sudo exposedwatch scan --external
```

- Objetivos por defecto: las direcciones IPv4 propias de este host que no sean de loopback ni de enlace local. No se escanea nada más a menos que lo indiques.
- Otros objetivos (`--external-target`; separados por comas o repetibles) deben ser direcciones IPv4, CIDR IPv4 (`/24` o más estrecho) o nombres de host. Los objetivos IPv6 aún no están soportados. Requieren `--confirm-ownership`. **Escanear sistemas que no son tuyos o sin permiso para hacerlo puede ser ilegal.**
- Solo escaneo TCP connect (`-sT`); no se escanea UDP. Usa `--external-ports 22,80,443,8000-8100` para elegir puertos (por defecto: los 1000 más comunes de nmap) y `--external-timeout` para limitar la duración (2m por defecto).
- nmap debe estar instalado (el `.deb` solo lo recomienda). Si falta, la fuente externa se informa como error y los resultados del host y de Docker se muestran igualmente.

## Instalación

### Ubuntu/Debian (`.deb`, recomendado)

Consulta el [Inicio rápido](#inicio-rápido). Verifica la descarga (opcional):

```sh
curl -fLO https://github.com/Byroncl/exposedwatch/releases/latest/download/checksums.txt
sha256sum --ignore-missing -c checksums.txt
```

Las sumas de verificación están en `checksums.txt`; ejecuta la comprobación en el directorio donde descargaste los archivos.

### Tarball

```sh
curl -fLO "https://github.com/Byroncl/exposedwatch/releases/latest/download/exposedwatch_linux_$(dpkg --print-architecture).tar.gz"
tar -xzf "exposedwatch_linux_$(dpkg --print-architecture).tar.gz" exposedwatch
sudo install -m 0755 exposedwatch /usr/local/bin/exposedwatch
```

### Con Go

```sh
go install github.com/Byroncl/exposedwatch/cmd/exposedwatch@latest
```

### Desinstalar

```sh
sudo apt remove exposedwatch
```

## Notas

- Ejecuta con `sudo` para poder resolver los nombres de proceso; sin él, los puertos se listan igualmente, pero la columna de proceso puede aparecer vacía.
- El descubrimiento de Docker lee el socket de Docker (`/var/run/docker.sock` por defecto, se cambia con `--docker-socket`). Tu usuario necesita acceso a él; de lo contrario, Docker se informa como una fuente que no se pudo leer.
- Solo lectura: lee `/proc` y envía una única solicitud `GET` al socket de Docker para listar los contenedores. No abre ni modifica puertos, ni modifica contenedores. Con `--external`, nmap además abre conexiones TCP hacia los objetivos escaneados.
- `--nmap-path` ejecuta el binario indicado con los mismos privilegios que exposedwatch, por lo que no permitas que usuarios no confiables pasen argumentos a `sudo exposedwatch` mediante reglas de sudoers.

## Licencia

Apache-2.0. Consulta [LICENSE](LICENSE).
