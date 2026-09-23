# 🚀 Guía Completa de Despliegue a Producción en VPS con Dokploy

Esta guía detalla el paso a paso para desplegar el backend de **Conexiate** en un servidor VPS propio utilizando **Dokploy** (orquestador Docker con proxy inverso Traefik y certificados SSL automáticos Let's Encrypt), y la compilación productiva del frontend en Flutter.

---

## 📋 Arquitectura del Stack en Dokploy

```
[ Internet ] ── HTTPS (443) ──> [ Traefik Proxy en VPS (Dokploy) ]
                                          │
                  ┌───────────────────────┴───────────────────────┐
                  ▼                                               ▼
     [ Backend Go (API :8080) ]                       [ Frontend Web (App) ]
                  │                                (Cloudflare / Dokploy Nginx)
                  ├────> [ Volumen Persistente: /app/uploads ]
                  │
                  ▼
   [ PostgreSQL Database (Dokploy / RDS) ]
```

---

## 🛠️ Requisitos Previos

1. **VPS configurado** (Ubuntu 22.04 o 24.04 LTS recomendado) con Dokploy instalado.
2. **IP pública del VPS**.
3. **Dominio configurado en tu proveedor DNS** (ej. Cloudflare, Namecheap, GoDaddy):
   * Registro `A`: `api.tudominio.com` ➔ Apuntando a la IP pública de tu VPS.
   * Registro `A`: `app.tudominio.com` ➔ Apuntando al hosting de tu frontend.

---

## Paso 1: Configurar la Base de Datos PostgreSQL en Dokploy

1. En el panel de **Dokploy**, ve al menú lateral izquierdo y haz clic en **Databases**.
2. Haz clic en **Create Database** y selecciona **PostgreSQL**.
3. Asigna los datos iniciales:
   * **Name**: `conexiate-postgres`
   * **Database Name**: `conexiate_prod`
   * **Username**: `postgres` (o el usuario que prefieras)
   * **Password**: Genera una contraseña segura y guárdala.
4. Haz clic en **Deploy**.
5. En la pestaña **General / Credentials** de la base de datos, copia:
   * El nombre del host interno (usualmente `conexiate-postgres` dentro de la red interna de Docker).
   * El puerto interno (`5432`).

---

## Paso 2: Crear el Servicio de Aplicación para el Backend

1. En Dokploy, ve a **Applications** ➔ **Create Application**.
2. Asigna un nombre: `conexiate-backend`.
3. En la sección **Source**:
   * **Source Type**: `Git`.
   * Conecta tu proveedor Git (GitHub, GitLab o Git genérico mediante SSH).
   * **Repository**: Selecciona el repositorio de Conexiate.
   * **Branch**: `main` (o tu rama de producción).
4. En la sección **Build Configuration**:
   * **Build Type**: Selecciona **Dockerfile**.
   * **Dockerfile Path**: `BASE_PROYECT_MULTI_ALL_BE/Dockerfile`
   * **Context Path**: `BASE_PROYECT_MULTI_ALL_BE`

---

## Paso 3: Configurar el Volumen Persistente (`uploads`)

> [!IMPORTANT]
> Es fundamental configurar este volumen para que las fotos de perfil, imágenes de beneficios y logos de los negocios aliados **no se eliminen** al actualizar la aplicación o reiniciar el contenedor.

1. Dentro de la aplicación `conexiate-backend` en Dokploy, ve a la pestaña **Volumes** (o **Mounts**).
2. Haz clic en **Add Volume**:
   * **Type**: `Bind` o `Volume` (recomendado: `Bind`).
   * **Host Path**: `/etc/dokploy/volumes/conexiate-uploads` (Dokploy creará esta ruta en el VPS).
   * **Container Path (Mount Path)**: `/app/uploads`
3. Guarda los cambios.

---

## Paso 4: Configurar Variables de Entorno en Dokploy

Ve a la pestaña **Environment** de tu aplicación en Dokploy y pega la siguiente configuración, reemplazando los valores reales:

```env
# ── Servidor y Rendimiento ──
GIN_MODE=release
SERVER_PORT=8080

# ── Base de Datos PostgreSQL ──
DB_HOST=conexiate-postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=TU_PASSWORD_SEGURO_DE_POSTGRES
DB_NAME=conexiate_prod
DB_SSLMODE=disable

# ── Autenticación JWT ──
# Generar una clave de 64 caracteres en consola: openssl rand -hex 32
JWT_SECRET=b68f4e2c1a89d70e45f31920bca3821094f6e8d2c1b50937a4e6f183920c5d71
JWT_EXPIRATION_HOURS=24

# ── Pasarela de Pagos Wompi (Modo Producción) ──
# Obtén estas llaves en https://comercios.wompi.co en modo Producción
WOMPI_PUBLIC_KEY=pub_prod_TU_LLAVE_PUBLICA
WOMPI_PRIVATE_KEY=prv_prod_TU_LLAVE_PRIVADA
WOMPI_EVENTS_SECRET=prod_events_TU_SECRETO_DE_EVENTOS
WOMPI_INTEGRITY_SECRET=prod_integrity_TU_SECRETO_DE_INTEGRIDAD
WOMPI_SANDBOX=false

# ── Servicio de Correo Transaccional (Resend) ──
RESEND_API_KEY=re_TU_LLAVE_API_DE_RESEND
RESEND_FROM_EMAIL=notificaciones@tudominio.com

# ── URL Pública del Frontend ──
FRONTEND_URL=https://app.tudominio.com
```

---

## Paso 5: Asignar Dominio y Certificado SSL (HTTPS)

1. En la aplicación `conexiate-backend`, ve a la pestaña **Domains**.
2. Haz clic en **Add Domain**:
   * **Host**: `api.tudominio.com`
   * **Path**: `/`
   * **Container Port**: `8080`
   * **HTTPS**: Marca la casilla **Enable HTTPS / Let's Encrypt**.
   * **Certificate Resolver**: `letsencrypt`
3. Guarda los cambios. Traefik generará y renovará automáticamente el certificado SSL gratuito.

---

## Paso 6: Desplegar y Validar el Backend

1. Haz clic en el botón superior **Deploy**.
2. Dirígete a la pestaña **Logs / Deployments** para observar el proceso:
   * Verás la descarga y compilación del binario en Alpine Linux.
   * Al arrancar, el backend ejecutará automáticamente las migraciones (`✅ Database migrations completed`) y la inicialización de roles y permisos (`✅ Database seeding completed`).
3. Prueba el estado del servidor desde tu terminal local o navegador:
   ```bash
   curl -I https://api.tudominio.com/api/health
   ```
   Debe responder `HTTP/2 200` con `{"status":"ok"}`.

---

## Paso 7: Compilar y Desplegar el Frontend (Flutter Web)

El frontend ya cuenta con soporte dinámico de la variable `API_URL` a través de `--dart-define`.

### 1. Compilación para Producción
En tu terminal local, dentro del directorio `BASE_PROYECT_MULTI_ALL_FE`, ejecuta:

```bash
flutter build web --release --dart-define=API_URL=https://api.tudominio.com/api
```

Los archivos finales listos para servir se generarán en:
`BASE_PROYECT_MULTI_ALL_FE/build/web/`

### 2. Dónde alojar el Frontend Web:
Puedes desplegar esta carpeta en cualquiera de las siguientes opciones recomendadas:
* **Cloudflare Pages / Vercel** (Gratuito, ultra rápido con CDN global y SSL automático).
* **En el mismo VPS con Dokploy**:
  * Creando una aplicación de tipo **Static** o **Docker (Nginx)** en Dokploy.
  * Dominio: `app.tudominio.com`.

---

## 🔒 Consideraciones de Seguridad y Mantenimiento

1. **Backups Automáticos de Base de Datos**:
   * En Dokploy, entra a tu base de datos `conexiate-postgres` ➔ pestaña **Backups**.
   * Configura una copia diaria automática (puedes guardarla localmente en el VPS o enviarla a AWS S3 / Cloudflare R2).
2. **Webhooks de Wompi en Producción**:
   * En el panel de comercios de Wompi, configura la URL de eventos apuntando a:
     `https://api.tudominio.com/api/membresia/webhook-wompi`
3. **Monitoreo de Recursos**:
   * En Dokploy, pestaña **Monitoring**, podrás ver el uso de CPU y memoria RAM en tiempo real.
