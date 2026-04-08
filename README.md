# Stream Engine MVP (OBS -> Go -> Browser)

Prototipo incremental para recibir video desde OBS por RTMP y reproducirlo en navegador por WebRTC.

## Arquitectura MVP

- OBS publica a MediaMTX por RTMP usando `stream_key` en la ruta.
- MediaMTX consulta al backend Go para autorizar publish/read.
- Navegador solicita al backend la URL WHEP y reproduce por WebRTC.

## Requisitos

- Go 1.22+
- Docker y Docker Compose
- OBS Studio

## Configuracion

1. Copia variables:

```bash
cp .env.example .env
```

2. Ajusta `STREAM_KEY` en `.env`.

## Ejecutar

1. Inicia MediaMTX:

```bash
docker compose up -d
```

2. Exporta variables y ejecuta backend:

```bash
set -a; source .env; set +a
go run ./cmd/server
```

3. Abre demo:

- [http://localhost:8080](http://localhost:8080)
- En el input usa: `/live/<tu_stream_key>`

### Modo ngrok (configuracion separada)

Esta modalidad no reemplaza la local; solo aplica cuando quieras exponer demo remota.

1. Crea env de ngrok:

```bash
cp .env.ngrok.example .env.ngrok
```

2. Levanta MediaMTX con config ngrok:

```bash
docker compose -f docker-compose.yml -f docker-compose.ngrok.yml up -d
```

3. Crea dos tuneles:

```bash
cp ./ngrok.yml.example ./ngrok.yml
ngrok start --all --config ./ngrok.yml
```

4. Actualiza `configs/mediamtx.ngrok.yml`:
- Cambia `REPLACE_WITH_8889_NGROK_HOST` por el dominio del tunel de `8889` (sin `https://`).

5. Actualiza `.env.ngrok`:
- Cambia `MEDIAMTX_HTTP_URL` por `https://<dominio-ngrok-8889>`.

6. Reinicia MediaMTX y backend:

```bash
docker compose -f docker-compose.yml -f docker-compose.ngrok.yml restart mediamtx
set -a; source .env.ngrok; set +a
go run ./cmd/server
```

7. Abre la URL publica de ngrok para `8080` y reproduce `/live/<tu_stream_key>`.

Nota: para redes estrictas/NAT compleja, agrega TURN en `configs/mediamtx.ngrok.yml`.

**Fallo común:** en `webrtcAdditionalHosts` debe ir **solo el hostname** (ej. `beff-123.ngrok-free.app`). Si pones `https://...`, MediaMTX intenta resolver literalmente `https://beff-...` y falla el WebRTC (`lookup https://...: no such host` en logs).

**Un solo agente ngrok:** si ves `ERR_NGROK_108`, cierra otros `ngrok` (`pgrep -a ngrok`; usa solo `ngrok start --all --config ./ngrok.yml`).

## Configurar OBS

- **Service**: Custom
- **Server**: `rtmp://localhost:1935/live`
- **Stream Key**: `<tu_stream_key>`

Cuando inicies streaming en OBS, la demo web debe mostrar el mismo video.

## Endpoints utiles

- `POST /auth/mediamtx`: hook de autorizacion para MediaMTX.
- `GET /api/viewer-session?path=/live/<key>`: retorna URL WHEP para el navegador.
- `GET /api/stats`: viewers conectados (contador simple en memoria).

## Notas de crecimiento

- Cambiar stream key estatica por login/JWT y llaves por canal.
- Persistir sesiones y metadata en DB/Redis.
- Agregar TURN server para redes restringidas.
- Escalar con nodos media separados y coordinacion de sesiones.
