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
sudo exposedwatch --no-connections     # desactiva la pestaña Connections
sudo exposedwatch --theme light        # fuerza la paleta para terminales claras
```

Opciones: `--proc-root`, `--docker-socket`, `--no-docker`, `--interval`, `--no-connections`, `--theme auto|dark|light` (por defecto `auto`, que detecta el fondo de la terminal). Requiere una terminal interactiva; de lo contrario termina con código 2 (usa `scan`). Si una fuente falla (por ejemplo, Docker no disponible), se muestra como una línea de advertencia y la vista continúa.

La cabecera muestra los conteos por severidad, un indicador de riesgo global (`SAFE`, `LOW`, `MEDIUM`, `HIGH` o `CRITICAL`, según la mayor severidad presente), el número de puertos por fuente y los contadores `+N new -M closed`. Cuando el ancho lo permite, la cabecera también muestra `conn N (in X / out Y)`. Cuatro pestañas comparten la pantalla:

- **Findings** (`1`): una fila por hallazgo.
- **All ports** (`2`): todos los puertos, con su severidad o `-` cuando ninguna regla aplica.
- **Changes** (`3`): registro de los puertos que se abrieron o cerraron desde que arrancó la vista, del más reciente al más antiguo (se conservan las últimas 200 entradas).
- **Connections** (`4`): quién está conectado a este servidor en este momento (ver más abajo).

El primer escaneo exitoso es la línea base: no se marca nada. Tras cada actualización posterior, los puertos que se abrieron reciben la marca `NEW` durante 5 actualizaciones y una entrada en el registro; los que se cerraron generan una entrada con su última severidad conocida. `+N new -M closed` cuenta estos eventos desde que arrancó la vista. Una actualización fallida, o una fuente que falla durante una actualización, nunca genera eventos `closed`: se conservan los datos previos de esa fuente. El registro vive solo en memoria y se pierde al salir. Un panel de detalle muestra la regla, la recomendación, la dirección de enlace, el proceso o contenedor y cuándo vio esta sesión por primera vez el puerto seleccionado.

| Tecla | Acción |
| ----- | ------ |
| `tab` / `shift+tab` | Pestaña siguiente / anterior |
| `1` / `2` / `3` / `4` | Ir a Findings / All ports / Changes / Connections |
| `up` / `k`, `down` / `j` | Mover la selección |
| `g` / `G` | Ir al inicio / al final |
| `r` | Actualizar ahora |
| `s` | Cambiar el orden: severidad, puerto, proceso (Findings y All ports); dirección, puerto local, dirección remota, proceso, antigüedad (Connections) |
| `f` | Cambiar el filtro de severidad mínima (Findings y All ports; se ignora en Connections) |
| `/` | Filtro de texto libre (`enter` aplica, `esc` cancela; en Connections busca en dirección remota, puerto remoto y proceso) |
| `l` | Mostrar u ocultar las conexiones de loopback a loopback (Connections; ocultas por defecto) |
| `i` | Mostrar solo las conexiones entrantes, o todas (Connections; por defecto se muestran todas) |
| `?` | Mostrar u ocultar la ayuda |
| `q` / `ctrl+c` | Salir |

### Pestaña Connections

La pestaña Connections lista las conexiones TCP establecidas de este equipo, leídas de `/proc/net/tcp` y `/proc/net/tcp6` en cada actualización. Cada fila muestra la dirección (`IN` cuando el puerto local es uno en el que este equipo escucha, `OUT` en caso contrario), el puerto local (y la dirección de enlace cuando hay espacio), la `dirección:puerto` remota, el estado, el proceso (`-` si se desconoce) y la antigüedad: cuánto tiempo lleva esta sesión viendo esa conexión (`12s`, `4m`, `1h`, `2d`). Las conexiones entrantes que aparecieron desde la actualización anterior reciben la marca `NEW` durante 5 actualizaciones; la primera lectura es la línea base y no marca nada. Las filas se ordenan primero las entrantes y luego por puerto local y dirección remota.

Las conexiones de loopback a loopback (servicios locales que hablan entre sí) están ocultas por defecto porque son ruido; pulsa `l` para mostrarlas. El título y el pie indican si están ocultas. Bajo la tabla, dos paneles listan las direcciones remotas más activas y el número de conexiones entrantes por puerto local, calculados con las mismas conexiones que la tabla (por eso siguen a `l` e `i`). Una lectura fallida conserva las filas anteriores y muestra una línea de advertencia; nunca retrasa el escaneo de puertos, que se actualiza de forma independiente. Si el número de conexiones alcanza el límite del lector, un aviso indica cuántas se leyeron.

Pulsa `i` para dejar solo las conexiones entrantes (`IN`); el título y el pie muestran `inbound:only` o `inbound:all`. Se combina con `l`, con el filtro `/` y con todos los órdenes. La tabla, el contador de la pestaña y el contador de la cabecera `conn N (in X / out Y)` describen lo que está visible, así que con `i` activo el contador muestra `out 0`; los dos paneles ya listan solo tráfico entrante. La selección se mantiene en la misma fila mientras esa fila siga visible.

Todo permanece en memoria: las direcciones remotas solo se muestran en pantalla, nunca se escriben en disco ni se registran, y no se resuelven por DNS.

#### Cómo interpretarla

Una conexión, aquí, es una sesión TCP abierta entre este equipo y otro. Solo se listan las sesiones `established`, es decir, conversaciones que están ocurriendo ahora mismo.

| Dirección | Qué significa | ¿Importa para la seguridad? |
| --------- | ------------- | --------------------------- |
| `IN` | Otro equipo se conectó a un puerto en el que este equipo escucha. | Sí, es el tráfico que hay que revisar: así llegan otros equipos a tus servicios. |
| `OUT` | Este equipo se conectó al servicio de otro. | Normalmente no por sí sola. Muestra con qué hablan tus programas. |

- Es normal ver muchas filas `OUT` en un equipo de escritorio: el navegador, el chat, la sincronización y las actualizaciones mantienen varias conexiones abiertas. Sus direcciones remotas suelen pertenecer a grandes proveedores de nube o CDN. Se muestran como números y no como nombres porque exposedwatch no hace consultas DNS a propósito.
- Fíjate primero en las filas `IN` que vienen de direcciones públicas hacia puertos que no esperabas, por ejemplo una base de datos o SSH.
- Consejos: `i` muestra solo las entrantes, `s` cambia el orden, `/` filtra por proceso o dirección, `l` muestra el tráfico de loopback y hace falta `sudo` para ver los nombres de proceso.

Limitaciones:

- Las conexiones se muestrean en cada actualización, por lo que las conexiones de muy corta duración entre dos actualizaciones pueden no verse.
- Solo se leen los sockets propios del equipo. Las conexiones a puertos publicados por Docker que el kernel gestiona mediante NAT pueden no aparecer.
- Los nombres de proceso requieren `sudo` (root), igual que la columna de proceso de las pestañas de puertos.
- Solo se listan conexiones TCP establecidas; UDP no se incluye.

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
