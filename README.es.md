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
| `?` | Mostrar u ocultar la ayuda |
| `q` / `ctrl+c` | Salir |

### Pestaña Connections

La pestaña Connections lista las conexiones TCP establecidas de este equipo, leídas de `/proc/net/tcp` y `/proc/net/tcp6` en cada actualización. Cada fila muestra la dirección (`IN` cuando el puerto local es uno en el que este equipo escucha, `OUT` en caso contrario), el puerto local (y la dirección de enlace cuando hay espacio), la `dirección:puerto` remota, el estado, el proceso (`-` si se desconoce) y la antigüedad: cuánto tiempo lleva esta sesión viendo esa conexión (`12s`, `4m`, `1h`, `2d`). Las conexiones entrantes que aparecieron desde la actualización anterior reciben la marca `NEW` durante 5 actualizaciones; la primera lectura es la línea base y no marca nada. Las filas se ordenan primero las entrantes y luego por puerto local y dirección remota.

Las conexiones de loopback a loopback (servicios locales que hablan entre sí) están ocultas por defecto porque son ruido; pulsa `l` para mostrarlas. El título y el pie indican si están ocultas. Bajo la tabla, dos paneles listan las direcciones remotas más activas y el número de conexiones entrantes por puerto local, calculados con las mismas conexiones que la tabla (por eso siguen a `l`). Una lectura fallida conserva las filas anteriores y muestra una línea de advertencia; nunca retrasa el escaneo de puertos, que se actualiza de forma independiente. Si el número de conexiones alcanza el límite del lector, un aviso indica cuántas se leyeron.

Todo permanece en memoria: las direcciones remotas solo se muestran en pantalla, nunca se escriben en disco ni se registran, y no se resuelven por DNS.

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

## Alertas (Telegram)

`exposedwatch watch` se ejecuta sin interfaz (no necesita terminal), escanea cada cierto intervalo y envía un mensaje de Telegram cuando aparece un puerto riesgoso **nuevo**. Está desactivado hasta que lo configures.

**Privacidad: léelo primero.**

- Las alertas se envían a los servidores de Telegram (`api.telegram.org`), un tercero. Cada mensaje contiene el nombre de este equipo, el puerto y el protocolo, la severidad, el nombre del proceso, el identificador de la regla y una recomendación breve.
- Las alertas nunca contienen direcciones IP remotas: describen lo que este equipo expone, no quién se conecta a él.
- Es la única función que realiza una conexión saliente, además del escaneo opcional con nmap. `scan` y la vista interactiva no envían nada a ningún lugar.
- El token del bot es un secreto. La herramienta lo muestra solo enmascarado (por ejemplo `123456789:****`) y lo elimina de los errores y logs que genera, y nunca lo escribe por sí misma. Mantén el archivo donde lo guardes con permisos `0600` o, mejor aún, mantén el token fuera del disco en una variable de entorno.
- Se respetan las variables de entorno estándar de proxy (`HTTPS_PROXY`, `NO_PROXY`). HTTPS pasa por el proxy mediante `CONNECT`, por lo que el proxy ve el host de destino (`api.telegram.org`) pero no la ruta de la solicitud, que contiene el token del bot.

### Configuración inicial

1. Crea un bot: abre un chat con [@BotFather](https://t.me/BotFather) en Telegram, envía `/newbot` y responde sus preguntas. BotFather te entrega el token del bot.
2. Obtén el id del chat: envía cualquier mensaje a tu nuevo bot (para un grupo, añade el bot al grupo y escribe allí). Luego abre `https://api.telegram.org/bot<TOKEN>/getUpdates` con tu token y busca `"chat":{"id":...}` en la respuesta. Los ids de grupo son números negativos.
3. Coloca los valores en el archivo de configuración o en el entorno y compruébalos:

```sh
exposedwatch notify-test
```

`notify-test` envía un mensaje de prueba e imprime una línea de éxito o el fallo. Muestra el token enmascarado y si proviene del archivo o del entorno, nunca el token completo. Códigos de salida: 0 correcto, 1 error de ejecución, 2 error de uso.

### Configuración

El archivo de configuración es JSON. Su ubicación por defecto es `/etc/exposedwatch/config.json` si ejecutas como root, y `$XDG_CONFIG_HOME/exposedwatch/config.json` (o `~/.config/exposedwatch/config.json`) en los demás casos. `--config RUTA` la reemplaza.

Todavía no hay ningún comando que escriba este archivo: lo creas y lo editas a mano. Como root, créalo con acceso privado antes de poner cualquier secreto en él:

```sh
sudo install -d -m 700 /etc/exposedwatch
sudo install -m 600 /dev/null /etc/exposedwatch/config.json
sudo $EDITOR /etc/exposedwatch/config.json
```

```json
{
  "version": 1,
  "telegram": { "enabled": true, "bot_token": "<token>", "chat_id": "<id del chat>" },
  "alerts": { "min_severity": "high", "cooldown_minutes": 60, "interval_seconds": 60 }
}
```

- Usa permisos `0600` para el archivo. Un permiso más abierto solo produce una advertencia. Se rechaza un enlace simbólico o un archivo que no pertenezca al usuario que ejecuta la herramienta. Como root, también se rechaza un archivo cuyo directorio puedan escribir el grupo u otros usuarios; los demás usuarios solo reciben una advertencia en ese caso.
- `EXPOSEDWATCH_TELEGRAM_TOKEN` y `EXPOSEDWATCH_TELEGRAM_CHAT_ID` tienen prioridad sobre el archivo, de modo que el token no tiene que estar en disco. Las alertas se envían solo cuando `telegram.enabled` es `true` y ambos valores están definidos en cualquiera de los dos lugares.
- `min_severity` es `info`, `low`, `medium`, `high` o `critical`. `interval_seconds` debe ser al menos 10.

### Watch

```sh
exposedwatch watch                 # se ejecuta hasta recibir SIGINT o SIGTERM
exposedwatch watch --once          # un solo escaneo y termina (para cron)
exposedwatch watch --interval 5m --min-severity critical
```

- El primer escaneo de cada ejecución envía un único resumen ("N findings at or above high already present", con un máximo de 5 títulos) y recuerda los hallazgos existentes sin alertar cada uno. Después, cada hallazgo que no era conocido alerta una vez. Un hallazgo que desaparece y reaparece dentro de `cooldown_minutes` no vuelve a alertar. Una fuente que falla durante un escaneo (por ejemplo Docker) no hace que sus hallazgos parezcan cerrados.
- Un mensaje que no se pudo entregar se reintenta con espera exponencial (comienza en el intervalo de escaneo y se duplica hasta 15 minutos) durante un máximo de 24 horas y luego se descarta; una alerta descartada permite que su hallazgo vuelva a alertar. Un envío fallido nunca detiene el bucle ni bloquea las alertas más nuevas. Los logs van a stderr, una línea con marca de tiempo por evento, sin tokens, URLs ni direcciones remotas.
- Una fuente que no se pudo leer durante el primer escaneo (por ejemplo, Docker aún sin iniciar) se resume una sola vez ("N existing findings from docker now visible") cuando pasa a ser legible, en lugar de alertar cada hallazgo.
- El estado se mantiene solo en memoria. `--once` no conserva estado entre ejecuciones, por lo que cada ejecución de cron vuelve a enviar el resumen inicial de lo que esté abierto en ese momento.
- Solo necesita root para lo que necesita el escaneo (nombres de proceso, socket de Docker).

Ejemplo de unidad systemd. **Es un ejemplo no probado, solo como documentación** (este proyecto no incluye empaquetado para ello). Se ejecuta como root. Guarda el secreto en un archivo legible solo por root (`chmod 0600`) y pon `"enabled": true` en el archivo de configuración:

```ini
# /etc/systemd/system/exposedwatch.service
[Unit]
Description=exposedwatch Telegram alerts
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/bin/exposedwatch watch
EnvironmentFile=/etc/exposedwatch/telegram.env
Restart=on-failure
NoNewPrivileges=yes
PrivateTmp=yes
ProtectHome=yes

[Install]
WantedBy=multi-user.target
```

Solo se sugiere un endurecimiento conservador. Opciones de aislamiento más estrictas pueden romper el acceso a `/proc` (nombres de proceso), al socket de Docker o a iptables, así que añádelas solo después de probarlas.

```sh
# /etc/exposedwatch/telegram.env  (permisos 0600, propietario root)
EXPOSEDWATCH_TELEGRAM_TOKEN=<token>
EXPOSEDWATCH_TELEGRAM_CHAT_ID=<id del chat>
```

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
- Solo lectura: lee `/proc` y envía una única solicitud `GET` al socket de Docker para listar los contenedores. No abre ni modifica puertos, ni modifica contenedores. Con `--external`, nmap además abre conexiones TCP hacia los objetivos escaneados. Solo `watch` y `notify-test` realizan una conexión saliente propia: envían las alertas a Telegram (consulta [Alertas (Telegram)](#alertas-telegram)).
- `--nmap-path` ejecuta el binario indicado con los mismos privilegios que exposedwatch, por lo que no permitas que usuarios no confiables pasen argumentos a `sudo exposedwatch` mediante reglas de sudoers.

## Licencia

Apache-2.0. Consulta [LICENSE](LICENSE).
