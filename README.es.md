[English](README.md) | Español

# exposedwatch

[![CI](https://github.com/Byroncl/exposedwatch/actions/workflows/ci.yml/badge.svg)](https://github.com/Byroncl/exposedwatch/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Byroncl/exposedwatch)](https://github.com/Byroncl/exposedwatch/releases/latest)
[![License](https://img.shields.io/github/license/Byroncl/exposedwatch)](LICENSE)

> **Estado: beta (v0.x). Aún no ha sido auditado por un revisor de seguridad humano independiente.** Consulta [SECURITY.md](SECURITY.md) para reportar una vulnerabilidad.

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
| `exposedwatch baseline save` | Registra los puertos expuestos como baseline; se usa con `scan --baseline` (ver [Modo baseline](#modo-baseline-cron)) |
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
sudo exposedwatch --no-connections     # desactiva el módulo Connections
sudo exposedwatch --no-firewall        # desactiva el módulo Firewall (no se lee el firewall)
sudo exposedwatch --theme light        # fuerza la paleta para terminales claras
exposedwatch --config ~/alertas.json   # archivo de configuración de Settings y Alerts (ruta por defecto más abajo)
```

Opciones: `--proc-root`, `--docker-socket`, `--no-docker`, `--interval`, `--no-connections`, `--no-firewall`, `--theme auto|dark|light` (por defecto `auto`, que detecta el fondo de la terminal), `--config RUTA` (el archivo de configuración que leen y escriben Settings y Alerts; tiene el mismo significado que en `watch` y `notify-test`). Requiere una terminal interactiva; de lo contrario termina con código 2 (usa `scan`). Si una fuente falla (por ejemplo, Docker no disponible), se muestra como una línea de advertencia y la vista continúa.

La cabecera muestra los conteos por severidad, un indicador de riesgo global (`SAFE`, `LOW`, `MEDIUM`, `HIGH` o `CRITICAL`, según la mayor severidad presente), el número de puertos por fuente y los contadores `+N new -M closed`. Cuando el ancho lo permite, la cabecera también muestra `conn N (in X / out Y)`. En Overview la cabecera es solo el título, porque Overview ya muestra esos números.

La pantalla tiene a la izquierda una barra lateral con módulos y, a la derecha, las pestañas del módulo actual seguidas de su vista:

| Módulo | Pestañas (`1`, `2`, ...) | Qué muestra |
| ------ | ------------------------ | ----------- |
| Overview | Summary | Indicador de riesgo global, conteos por severidad, contadores, los 5 peores hallazgos, los últimos 5 cambios y las 3 direcciones remotas principales (entrantes, solo cuando las conexiones están activadas) |
| Ports | Findings, All ports, Changes | `Findings`: una fila por hallazgo. `All ports`: todos los puertos, con su severidad o `-` cuando ninguna regla aplica. `Changes`: registro de los puertos que se abrieron o cerraron desde que arrancó la vista, del más reciente al más antiguo (se conservan las últimas 200 entradas) |
| Connections | Live, Top remote | `Live`: quién está conectado a este servidor en este momento (ver más abajo). `Top remote`: las direcciones remotas entrantes más activas y el número de conexiones entrantes por puerto local, a tamaño completo |
| Firewall | Status, Ports | Qué puertos en escucha deja pasar el firewall del host, solo como información (ver [Veredictos del firewall](#veredictos-del-firewall)). `Status`: el backend, si se pudieron leer las reglas, las notas del motor, un conteo por veredicto, cuándo se evaluó por última vez y qué significa cada veredicto. `Ports`: una fila por puerto del host o de Docker con su veredicto, los más preocupantes primero |
| Alerts | Status, Preview | `Status`: si Telegram está configurado (token enmascarado; el id del chat solo se indica como definido), activado o desactivado, severidad mínima, enfriamiento, intervalo de revisión, archivo de configuración con sus advertencias y el resultado del último mensaje de prueba de esta sesión. `Preview`: los hallazgos actuales con severidad igual o superior a la mínima guardada, como "alertarían ahora". La TUI nunca envía alertas por sí misma: las envía `exposedwatch watch`, y tanto `notify-test` como la fila `[ Send test message ]` envían un mensaje de prueba a Telegram (consulta [Settings > Telegram y el módulo Alerts](#settings--telegram-y-el-módulo-alerts)) |
| Settings | Telegram, Display | `Telegram`: un formulario editable de la configuración de alertas (ver más abajo). `Display`: el tema, el intervalo de actualización y si las conexiones y el firewall están activados, en modo lectura |

La vista arranca en Ports, Findings. Cada módulo recuerda la última pestaña que usaste en él. Overview también muestra una línea `Alerts: on (min high)` o `Alerts: off (min high)` (`n/a` cuando no se puede cargar el archivo de configuración).

La barra lateral se adapta al ancho de la terminal: completa (nombres de módulo con un conteo o marca, por ejemplo el número de hallazgos y `NEW` en Ports, el número de conexiones en Connections, el número de puertos que el firewall permite o restringe en Firewall (`n/a` cuando no se pueden leer las reglas, `off` con `--no-firewall`) y, en Alerts, `on` si las alertas están activadas y configuradas, `off` si están desactivadas o sin configurar, y `n/a` si no se puede cargar el archivo de configuración) desde 80 columnas, compacta (una letra por módulo) de 50 a 79 columnas, y oculta por debajo de 50. `b` la oculta o la muestra además de lo anterior. Nada se dibuja fuera de la terminal, sea cual sea el tamaño.

El primer escaneo exitoso es la línea base: no se marca nada. Tras cada actualización posterior, los puertos que se abrieron reciben la marca `NEW` durante 5 actualizaciones y una entrada en el registro; los que se cerraron generan una entrada con su última severidad conocida. `+N new -M closed` cuenta estos eventos desde que arrancó la vista. Una actualización fallida, o una fuente que falla durante una actualización, nunca genera eventos `closed`: se conservan los datos previos de esa fuente. El registro vive solo en memoria y se pierde al salir. Un panel de detalle muestra la regla, la recomendación, la dirección de enlace, el proceso o contenedor y cuándo vio esta sesión por primera vez el puerto seleccionado.

| Tecla | Acción |
| ----- | ------ |
| `tab` / `shift+tab` | Módulo siguiente / anterior |
| `]` / `[` | Pestaña siguiente / anterior del módulo actual |
| `1` a `9` | Ir a esa pestaña del módulo actual (Ports: Findings / All ports / Changes; Connections: Live / Top remote; Firewall: Status / Ports; Alerts: Status / Preview; Settings: Telegram / Display) |
| `b` | Mostrar u ocultar la barra lateral |
| `up` / `k`, `down` / `j` | Mover la selección (tablas) |
| `g` / `G` | Ir al inicio / al final (tablas) |
| `enter` | Settings > Telegram: activar o alternar la fila seleccionada, empezar a editar un valor, o ejecutar `[ Save ]` / `[ Send test message ]` |
| mientras se edita | Settings > Telegram: escribe o pega el valor; `enter` guarda el valor, `esc` cancela, `backspace` borra. Ninguna otra tecla actúa, de modo que se puede escribir cualquier carácter (`ctrl+c` sigue saliendo) |
| `r` | Actualizar ahora (también vuelve a leer las reglas del firewall) |
| `s` | Cambiar el orden: severidad, puerto, proceso (Findings y All ports); dirección, puerto local, dirección remota, proceso, antigüedad (Connections / Live); veredicto, puerto, proceso (Firewall / Ports) |
| `f` | Cambiar el filtro de severidad mínima (solo Findings y All ports) |
| `/` | Filtro de texto libre (`enter` aplica, `esc` cancela; en Connections / Live busca en dirección remota, puerto remoto y proceso; en Firewall / Ports busca en puerto, proceso, veredicto y motivo) |
| `l` | Mostrar u ocultar las conexiones de loopback a loopback (Connections: Live y Top remote; ocultas por defecto) |
| `i` | Mostrar solo las conexiones entrantes, o todas (Connections: Live y Top remote; por defecto se muestran todas) |
| `?` | Mostrar u ocultar la ayuda |
| `q` / `ctrl+c` | Salir |

Las teclas que no se aplican a la pestaña actual se ignoran y sus indicaciones no aparecen en el pie. Cuando el pie es demasiado estrecho, se descartan primero las indicaciones menos importantes.

### Módulo Connections (pestaña Live)

La pestaña Live del módulo Connections lista las conexiones TCP establecidas de este equipo, leídas de `/proc/net/tcp` y `/proc/net/tcp6` en cada actualización. Cada fila muestra la dirección (`IN` cuando el puerto local es uno en el que este equipo escucha, `OUT` en caso contrario), el puerto local (y la dirección de enlace cuando hay espacio), la `dirección:puerto` remota, el estado, el proceso (`-` si se desconoce) y la antigüedad: cuánto tiempo lleva esta sesión viendo esa conexión (`12s`, `4m`, `1h`, `2d`). Las conexiones entrantes que aparecieron desde la actualización anterior reciben la marca `NEW` durante 5 actualizaciones; la primera lectura es la línea base y no marca nada. Las filas se ordenan primero las entrantes y luego por puerto local y dirección remota.

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

### Veredictos del firewall

El módulo Firewall, y una línea `Firewall:` en el detalle de las tablas de puertos, indican si el firewall del host dejaría pasar una conexión entrante NUEVA desde un origen externo hasta cada puerto en escucha: `allowed` (permitido), `restricted` (solo algunas direcciones de origen), `blocked` (bloqueado), `unknown` (no se pudo determinar), `not-exposed` (solo loopback) o `docker` (publicado por Docker, que normalmente salta las reglas INPUT y ufw, por lo que nunca se informa como bloqueado). Las tablas de Ports también tienen una columna `FW` de baja prioridad, la primera en descartarse cuando la terminal es estrecha, y Overview incluye un resumen de una línea. Los veredictos son **informativos**: nunca cambian una severidad, un hallazgo ni un código de salida.

Las reglas se leen en segundo plano con su propio límite de tiempo: tras el primer escaneo, con `r`, cuando cambia el conjunto de puertos en escucha y, en los demás casos, como máximo cada 30 segundos (`iptables-save` no es gratuito), con independencia del intervalo de actualización. Una lectura fallida conserva los datos anteriores y muestra una advertencia en las vistas de Firewall. Sin root el módulo queda en `n/a`: todos los puertos expuestos son `unknown` y las notas explican el motivo. `--no-firewall` desactiva el módulo y el firewall nunca se lee.

#### Limitaciones

- Requiere root: sin él no se pueden leer las reglas y todos los puertos expuestos son `unknown`.
- Lee `iptables-save` y `ip6tables-save` (en `/usr/sbin` y `/sbin`, que no están en el `PATH` de un usuario normal). Esto cubre ufw, iptables puro y la capa de compatibilidad iptables-nft. Los equipos que solo usan nftables nativo se informan como `unknown`; una `table ip filter` nativa creada a mano no se detecta.
- Los puertos publicados por Docker se etiquetan como `docker`, nunca como `blocked`.
- Un veredicto es una estimación para una conexión entrante nueva desde un origen externo. Todo lo que el motor no entiende deja el veredicto en `unknown`.
- Se validó únicamente con datos de prueba (fixtures), no con conjuntos de reglas reales en equipos reales.

## Scripts, cron y CI: `scan`

```sh
sudo exposedwatch scan                       # salida en tabla
sudo exposedwatch scan --format json         # salida legible por máquinas
sudo exposedwatch scan --min-severity medium # muestra y evalúa desde medium en adelante
sudo exposedwatch scan --no-docker           # omite el descubrimiento de Docker
sudo exposedwatch scan --firewall            # muestra también el veredicto del firewall de cada puerto
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

`--firewall` es opcional. Añade una columna `FIREWALL` con el veredicto de cada puerto y una breve sección `Firewall` con el backend, si se pudieron leer las reglas y las notas del motor; con `--format json` añade un objeto `firewall` de primer nivel (`available`, `backend`, `notes` y `ports` con `port`, `protocol`, `bind_address`, `source`, `verdict` y `reason`; en JSON el veredicto de Docker es `docker-published`). No poder leer el firewall (sin root, sin la herramienta) no es un error: la sección lo indica y todos los puertos expuestos son `unknown`. Los veredictos son informativos: las severidades y los códigos de salida son exactamente los mismos con y sin `--firewall`. Consulta las [limitaciones](#limitaciones).

Códigos de salida (comunes a todos los comandos):

| Código | Significado |
| ------ | ----------- |
| 0 | Correcto, sin hallazgos en `--min-severity` o superior |
| 1 | Error de ejecución |
| 2 | Error de uso (opción o comando inválido) |
| 3 | Se encontraron hallazgos en `--min-severity` o superior |
| 4 | Deriva: puertos nuevos respecto a una baseline (solo `scan --baseline`, que nunca usa el 3; ver [Modo baseline](#modo-baseline-cron)) |

Ejemplo para cron o CI: hacer fallar el trabajo si hay algo de severidad medium o mayor expuesto.

```sh
exposedwatch scan --min-severity medium --format json > exposure.json || echo "exposure check exited with $?"
```

## Modo baseline (cron)

Una baseline es la lista de puertos que has aceptado como expuestos. `scan --baseline` compara lo que está expuesto ahora con esa lista e informa solo la diferencia, de modo que un trabajo de cron permanece en silencio hasta que aparece algo nuevo. Ambos comandos son solo informativos: nunca abren, cierran ni modifican nada del sistema.

```sh
sudo exposedwatch baseline save              # registra los puertos expuestos ahora
sudo exposedwatch scan --baseline /var/lib/exposedwatch/baseline.json
```

`baseline save` ejecuta el escaneo del host y, salvo que se indique `--no-docker`, el de Docker (nunca el escaneo externo con nmap), escribe la baseline e imprime una línea: `Baseline saved to <ruta> (<N> ports)`. Opciones: `--file`, `--proc-root`, `--docker-socket`, `--no-docker`, `--allow-partial`, `--force`.

`scan --baseline ARCHIVO` funciona con `--format table|json`, `--no-docker`, `--proc-root`, `--docker-socket`, `--firewall`, `--external` y `--min-severity`. Imprime el informe normal más una sección `Drift` (o, con `--format json`, un objeto `drift` de nivel superior con `baseline_file`, `created_at`, `new`, `closed` y `notes`; las listas vacías son `[]`). Sin `--baseline` la salida es exactamente la misma de antes.

Códigos de salida con `--baseline`:

| Código | Significado |
| ------ | ----------- |
| 0 | Sin deriva: ningún puerto nuevo en `--min-severity` o superior (también cuando solo difieren puertos cerrados) |
| 1 | Error de ejecución, incluida una baseline ausente, no confiable, corrupta o ilegible (ejecuta `exposedwatch baseline save`), o un origen (host, Docker, externo) que no se pudo leer durante el escaneo sin que ningún puerto nuevo alcanzara `--min-severity` |
| 2 | Error de uso |
| 4 | Deriva: al menos un puerto NUEVO alcanza `--min-severity`; prevalece sobre el 1 cuando ambos aplican |

El código 3 **no** se usa en modo baseline, porque se espera que los puertos aceptados sigan expuestos. Un puerto nuevo cuenta con la severidad de su hallazgo más grave; un puerto sin hallazgos (por ejemplo, uno en loopback) cuenta como `info`, así que con el `--min-severity info` por defecto cualquier puerto nuevo cuenta. Los puertos que están en la baseline pero ya no están expuestos se listan como `CLOSED` y son informativos: una deriva solo de puertos cerrados termina con 0. Si un origen falla durante el escaneo, el resultado es parcial y cron nunca ve una ejecución limpia en silencio: stderr recibe una línea `unverified sources: <nombres>`, la sección Drift y las `notes` del JSON lo indican (siempre, tenga o no la baseline entradas de ese origen) y el código de salida es 1, salvo que se encuentren puertos nuevos en `--min-severity` o superior, en cuyo caso es 4.

Ejemplo de cron (adapta la ruta, la frecuencia y el comando de aviso a tu sistema):

```sh
# Solo un ejemplo: cada hora, ejecuta la comparación y reacciona cuando termina con un código distinto de cero.
0 * * * * root exposedwatch scan --baseline /var/lib/exposedwatch/baseline.json --no-docker >/dev/null || logger -t exposedwatch "deriva o error: código de salida $?"
```

Como un código distinto de cero también puede ser un error de ejecución (1), revisa el código de salida, y no solo el éxito, cuando necesites distinguir la deriva (4) de una baseline dañada.

Qué contiene la baseline y dónde se guarda:

- La identidad de un puerto es protocolo + puerto + dirección de enlace + origen (`host` o `docker`). Un servicio que pasa de `127.0.0.1` a `0.0.0.0` es un puerto NUEVO, porque cambió su exposición.
- El archivo es JSON con `version`, `created_at` (UTC), `tool_version` y las `entries`, ordenadas. No contiene nombres de procesos, nombres de contenedores, direcciones remotas ni el nombre del host, por lo que es seguro copiarlo.
- Ruta por defecto: `/var/lib/exposedwatch/baseline.json` como root; en otro caso `$XDG_STATE_HOME/exposedwatch/baseline.json` (si es una ruta absoluta) o `~/.local/state/exposedwatch/baseline.json`. `--file` la reemplaza. El directorio se crea con modo 0700 y el archivo con 0600, escrito de forma atómica (si el guardado falla, se conserva el archivo anterior).
- La carga es estricta, porque una baseline alterada ocultaría puertos nuevos. Se rechaza el archivo si es un enlace simbólico, si cambió mientras se abría, si no pertenece al usuario que ejecuta el comando (como root: si no pertenece a root) o si el grupo u otros pueden escribirlo; además debe ser un archivo regular (un FIFO, dispositivo o directorio se rechaza sin quedarse bloqueado). Como root también se rechaza si su directorio es escribible por el grupo u otros o no pertenece a root; como otro usuario solo se muestra una advertencia (el directorio debe pertenecer a ese usuario o a root). Solo se revisa el directorio inmediato, no sus ancestros, así que mantén el directorio de estado bajo ancestros que pertenezcan a root. Se leen hasta 4 MiB; se rechazan los campos desconocidos, una versión distinta de 1, las entradas inválidas y más de 20000 entradas, y las entradas duplicadas se unifican. Los mensajes de error nunca muestran el contenido del archivo.
- Si no se puede leer un origen al guardar (por ejemplo, Docker no está en ejecución), `baseline save` se niega a escribir una baseline parcial, que ocultaría todo ese origen; usa `--allow-partial` para guardarla de todos modos. Si no se puede leer un origen durante `scan --baseline`, sus puertos son desconocidos, no cerrados: sus entradas de la baseline no se informan como cerradas (ver los códigos de salida más arriba).
- `baseline save` se niega a sobrescribir una ruta existente que no sea una baseline: el destino debe no existir, o ser un archivo regular (nunca un enlace simbólico) que se cargue como baseline confiable y válida; en otro caso termina con 1 y el mensaje `refusing to overwrite <ruta>: it is not a baseline file (use --force to replace it anyway)`. `--force` es peligroso: reemplaza un archivo regular aunque su contenido no sea una baseline, y aun así nunca sigue un enlace simbólico ni reemplaza un directorio.
- Las direcciones de enlace se guardan en forma canónica (minúsculas, IPv6 mapeado a IPv4 escrito como IPv4, sin zona IPv6); `0.0.0.0` y `::` siguen siendo entradas distintas.

`exposedwatch watch --baseline ARCHIVO` usa la misma baseline en el bucle de alertas sin terminal: consulta [Alertas solo de deriva con baseline](#alertas-solo-de-deriva-con-baseline).

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
- Es la única función que realiza una conexión saliente, además del escaneo opcional con nmap. `scan` y la vista interactiva no envían nada a ningún lugar, salvo el mensaje de prueba que inicias tú desde Settings > Telegram (también va a Telegram).
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

Puedes editar este archivo a mano o desde la vista interactiva (consulta [Settings > Telegram y el módulo Alerts](#settings--telegram-y-el-módulo-alerts)). Si lo editas a mano como root, créalo con acceso privado antes de poner cualquier secreto en él:

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

### Settings > Telegram y el módulo Alerts

En la vista interactiva, Settings > Telegram edita el mismo archivo de configuración que leen `watch` y `notify-test` (`--config RUTA` elige otro):

| Fila | Cómo funciona |
| ---- | ------------- |
| Enabled | `enter` la activa o desactiva |
| Bot token | `enter` empieza a escribir (o pegar) un token nuevo; parte vacío y cada carácter escrito se muestra como un punto. Cuando no se edita, solo se muestra la forma enmascarada (por ejemplo `123456789:****`). Confirmar un valor vacío conserva el token actual: un token se puede cambiar pero no borrar desde el formulario (quítalo editando el archivo) |
| Chat ID | `enter` empieza a editar |
| Minimum severity | `enter` recorre info, low, medium, high, critical |
| Cooldown (minutes), Check interval (seconds) | `enter` empieza a editar; solo dígitos (los demás caracteres se ignoran y se muestra la indicación `digits only`) |
| `[ Save ]` | Escribe el archivo |
| `[ Send test message ]` | Envía un mensaje de prueba a Telegram con los valores que hay en el formulario, guardados o no, y muestra `Sending...` y luego `Sent` o el fallo. Requiere un token y un id de chat. Solo hay una prueba a la vez y la vista sigue funcionando mientras se envía (el intento completo se limita a 15 segundos) |

- Te mueves entre filas con `j`/`k` o las flechas. Mientras se escribe un valor, ninguna otra tecla se interpreta: `q`, `tab`, `b`, `?` y los dígitos son texto. `enter` conserva el valor y `esc` lo cancela. Pegar funciona (se descartan los caracteres de control). `ctrl+c` sigue saliendo.
- Se muestra `unsaved changes` hasta que guardes. Los límites de enfriamiento e intervalo, y cualquier otra validación, son los del archivo de configuración; un guardado rechazado muestra el motivo y deja el archivo intacto. `[ Save ]` no hace nada (`No changes to save`) si no hay cambios desde la última carga o guardado, y `[ Send test message ]` espera 5 segundos entre dos envíos.
- Guardar escribe el archivo con las mismas reglas de arriba: un archivo privado (`0600`), escrito de forma atómica, en un directorio que crea con `0700` si no existe; guardar además corrige los permisos más abiertos de un archivo anterior. No se sobrescribe un archivo que sea un enlace simbólico, que no sea tuyo o que no se pueda cargar, y el formulario queda en solo lectura mostrando el motivo.
- Guardar desde la vista interactiva reescribe el archivo solo con los ajustes conocidos: no se conservan los comentarios ni las claves desconocidas. Si mantienes claves adicionales, edita el archivo a mano. Como root, se rechaza guardar cuando el directorio de configuración puede ser escrito por el grupo u otros; como usuario normal se guarda y se muestra una advertencia.
- Si el token o el id del chat provienen de `EXPOSEDWATCH_TELEGRAM_TOKEN` / `EXPOSEDWATCH_TELEGRAM_CHAT_ID`, su fila es de solo lectura e indica `from environment (...)`. Los valores del entorno nunca se escriben en el archivo; al guardar se conserva lo que el archivo ya tuviera para ellos.
- `exposedwatch watch` lee el archivo al arrancar y no lo recarga: reinicia `watch` para aplicar un cambio guardado. El módulo Alerts muestra los valores guardados, no las ediciones sin guardar.
- Privacidad: un mensaje de prueba va a los servidores de Telegram como cualquier alerta (ver arriba). La vista nunca muestra el token, ni lo registra, ni lo incluye en un error; solo viaja al archivo de configuración cuando lo guardas y a Telegram cuando envías un mensaje de prueba.

### Watch

```sh
exposedwatch watch                 # se ejecuta hasta recibir SIGINT o SIGTERM
exposedwatch watch --once          # un solo escaneo y termina (para cron)
exposedwatch watch --interval 5m --min-severity critical
```

- El primer escaneo de cada ejecución envía un único resumen ("N findings at or above high already present", con un máximo de 5 títulos) y recuerda los hallazgos existentes sin alertar cada uno. Después, cada hallazgo que no era conocido alerta una vez. Un hallazgo que desaparece y reaparece dentro de `cooldown_minutes` no vuelve a alertar. Una fuente que falla durante un escaneo (por ejemplo Docker) no hace que sus hallazgos parezcan cerrados.
- Un mensaje que no se pudo entregar se reintenta con espera exponencial (comienza en el intervalo de escaneo y se duplica hasta 15 minutos) durante un máximo de 24 horas y luego se descarta; una alerta descartada permite que su hallazgo vuelva a alertar. Un envío fallido nunca detiene el bucle ni bloquea las alertas más nuevas. Los logs van a stderr, una línea con marca de tiempo por evento, sin tokens, URLs ni direcciones remotas.
- Una fuente que no se pudo leer durante el primer escaneo (por ejemplo, Docker aún sin iniciar) se resume una sola vez ("N existing findings from docker now visible") cuando pasa a ser legible, en lugar de alertar cada hallazgo.
- El estado se mantiene solo en memoria. Sin `--baseline`, `--once` no conserva estado entre ejecuciones, por lo que cada ejecución de cron vuelve a enviar el resumen inicial de lo que esté abierto en ese momento; usa una baseline para evitarlo (sección siguiente).
- Solo necesita root para lo que necesita el escaneo (nombres de proceso, socket de Docker).

### Alertas solo de deriva con baseline

`exposedwatch watch --baseline ARCHIVO` (también con `--once`) alerta únicamente sobre la deriva respecto a una baseline creada con `exposedwatch baseline save`:

- Un hallazgo está aceptado cuando la identidad de su puerto (protocolo, puerto, dirección de enlace canónica y fuente `host` o `docker`) figura en la baseline. Los hallazgos aceptados nunca alertan, ni de forma individual ni en un resumen. Los nombres de proceso se ignoran, y el mismo puerto en otra dirección de enlace o desde otra fuente no se considera aceptado.
- No hay resumen inicial. Cada hallazgo con severidad igual o superior a `--min-severity` (o `alerts.min_severity`) que no esté aceptado alerta de forma individual, con una línea adicional en el cuerpo: `Not in baseline (saved AAAA-MM-DD)`. El enfriamiento, la deduplicación, los reintentos y el tratamiento de las fuentes ilegibles funcionan como se describió antes. Los puertos cerrados nunca alertan.
- La baseline debe ser de confianza, con las mismas reglas que `scan --baseline` (no es un enlace simbólico, pertenece al usuario actual o a root, no es escribible por grupo ni otros, es un archivo regular, el directorio padre es seguro y el análisis es estricto y acotado). Si falta, no es de confianza o está corrupta, `watch` se niega a arrancar con código de salida 1 y una indicación para ejecutar `exposedwatch baseline save`; nunca continúa sin baseline. `--baseline ""` es un error de uso (código de salida 2).
- En modo bucle, el archivo se recarga al inicio de cada ciclo, de modo que `exposedwatch baseline save` surte efecto sin reiniciar `watch`: los hallazgos que pasan a estar aceptados se olvidan en silencio y los que dejan de estarlo alertan. Si una recarga falla tras una carga correcta, `watch` conserva la última baseline válida y registra una línea (`baseline reload failed: ...; using the last good baseline`), otra si cambia el motivo y otra cuando se recupera; no repite el mensaje en cada ciclo.
- `--once` carga la baseline una sola vez y no conserva estado, por lo que cada ejecución de cron alerta sobre la deriva presente en ese momento: la misma deriva vuelve a alertar en cada ejecución hasta que la aceptes actualizando la baseline con `baseline save`.
- Aceptar un puerto acepta TODOS los hallazgos de ese puerto, sea cual sea la regla o la severidad: un hallazgo más grave en un puerto ya aceptado NO alerta hasta que el puerto se elimine de la baseline. Una baseline vacía no acepta nada, así que alerta todo hallazgo con severidad igual o superior a la mínima.
- Una recarga que cambia el número de entradas o la fecha de guardado registra una línea (`baseline reloaded: N entries (was M), saved FECHA`); añade `WIDENED` cuando la baseline creció en al menos 50 entradas o más del doble. Las recargas sin cambios no registran nada. Un fallo se registra una vez por cada motivo distinto, y las advertencias de carga solo cuando cambian.
- Los logs nunca incluyen el contenido del archivo. La ruta aparece en la línea de arranque, y los mensajes de error pueden incluir la ruta que indicaste.

Ejemplo de unidad systemd. **Es un ejemplo no probado, solo como documentación** (este proyecto no incluye empaquetado para ello). Crea antes la baseline (`exposedwatch baseline save`); sin un archivo de baseline de confianza, `watch --baseline` se niega a arrancar. Se ejecuta como root. Guarda el secreto en un archivo legible solo por root (`chmod 0600`) y pon `"enabled": true` en el archivo de configuración:

```ini
# /etc/systemd/system/exposedwatch.service
[Unit]
Description=exposedwatch Telegram alerts
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/bin/exposedwatch watch --baseline /var/lib/exposedwatch/baseline.json
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

## Estado y qué está verificado

Verificado:

- Se instala y funciona desde el `.deb` publicado en Debian 12 y Ubuntu 22.04 y 24.04 (`amd64`) en contenedores limpios, y se desinstala sin dejar restos.
- Desarrollado y ejecutado en Debian 13.
- El código está cubierto por una suite de pruebas automatizadas, y las partes sensibles para la seguridad se revisaron con revisión adversarial automatizada. No ha habido una auditoría humana independiente.

Aún no verificado:

- El módulo de firewall con reglas reales como root (solo se ejercitaron datos de prueba y la ruta sin root).
- El envío real a Telegram.
- La visualización de la TUI en distintas terminales y colores.
- El uso con cron y systemd.
- El paquete `arm64`.
- Distribuciones distintas de Debian y Ubuntu.
- Las reglas nativas de nftables `table ip filter` hechas a mano no se detectan.

## Seguridad y uso responsable

- Puede ejecutarse como root (para nombres de proceso y reglas del firewall), con los privilegios del usuario que lo inicia.
- Solo lee el firewall: nunca modifica sus reglas.
- El escaneo opcional con nmap (`--external`) solo debe apuntar a sistemas que sean tuyos o que tengas permiso para probar.
- Los secretos no aparecen en los logs: el token del bot de Telegram se muestra enmascarado y se elimina de los errores y logs.
- Para reportar una vulnerabilidad, sigue [SECURITY.md](SECURITY.md). No abras un issue público.

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
- Solo lectura: lee `/proc` y envía una única solicitud `GET` al socket de Docker para listar los contenedores. No abre ni modifica puertos, ni modifica contenedores. Con `--external`, nmap además abre conexiones TCP hacia los objetivos escaneados. Solo `watch`, `notify-test` y el mensaje de prueba de Settings > Telegram realizan una conexión saliente propia: envían las alertas a Telegram (consulta [Alertas (Telegram)](#alertas-telegram)).
- `--nmap-path` ejecuta el binario indicado con los mismos privilegios que exposedwatch, por lo que no permitas que usuarios no confiables pasen argumentos a `sudo exposedwatch` mediante reglas de sudoers.

## Licencia

Apache-2.0. Consulta [LICENSE](LICENSE).
