# Guía de perfiles de distro

Bootimus usa perfiles de distro para detectar tipos de ISO y generar los parámetros de arranque correctos. Los perfiles son data-driven — puedes añadir soporte para nuevas distribuciones sin cambiar el código.

## Tabla de contenidos

- [Visión general](#visión-general)
- [Cómo funciona](#cómo-funciona)
- [Ver perfiles](#ver-perfiles)
- [Actualizar perfiles](#actualizar-perfiles)
- [Crear perfiles custom](#crear-perfiles-custom)
- [Campos del perfil](#campos-del-perfil)
- [Placeholders](#placeholders)
- [Ejemplos](#ejemplos)
- [Solución de problemas](#solución-de-problemas)

## Visión general

Los perfiles de distro definen:
- **Cómo detectar** qué distro es un ISO (match de patrones de filename)
- **Dónde encontrar** el kernel, initrd y squashfs dentro del ISO
- **Qué parámetros de arranque** usar al arrancar por PXE
- **Qué tipo de auto-instalación** se soporta (preseed, kickstart, autoinstall, etc.)

### Tipos de perfil

| Tipo | Descripción |
|------|-------------|
| **Built-in** | Incluido con Bootimus, actualizado desde el repositorio central |
| **Custom** | Creado por el usuario, nunca sobrescrito por actualizaciones |

Los perfiles custom siempre tienen prioridad sobre los built-in al hacer match de nombres de ISO.

## Cómo funciona

1. Cuando un ISO se sube o se extrae, Bootimus hace match del filename contra los patrones de perfil
2. Los paths de kernel/initrd del perfil coincidente se usan para localizar archivos de arranque dentro del ISO
3. Los boot params del perfil se convierten en el default (editable en las Properties de la imagen)
4. En el momento del arranque, los placeholders en los params se resuelven a URLs reales

### Ciclo de vida del perfil

```
Build time:    distro-profiles.json embedded in binary
                        ↓
First startup:  Profiles seeded into database
                        ↓
"Check for Updates":  Latest profiles fetched from GitHub
                        ↓
User creates:   Custom profiles stored in database (never overwritten)
```

## Ver perfiles

Navega a **Boot > Distro Profiles** en el panel admin para ver todos los perfiles cargados con sus patrones de filename, parámetros de arranque, tipo (Built-in/Custom) y versión.

## Actualizar perfiles

### Automático (recomendado)

Haz click en **"Check for Updates"** en la pestaña Distro Profiles. Esto descarga los últimos perfiles desde:

```
https://raw.githubusercontent.com/garybowers/bootimus/main/distro-profiles.json
```

- Los perfiles nuevos se añaden automáticamente
- Los perfiles built-in existentes se actualizan a la última versión
- Los perfiles custom nunca se modifican

### Vía API

```bash
curl -H "Authorization: Bearer $TOKEN" -X POST http://localhost:8081/api/profiles/update
```

Respuesta:
```json
{
  "success": true,
  "message": "Updated to version 0.1.21 (2 added, 5 updated)"
}
```

## Crear perfiles custom

### Vía interfaz web

1. Ve a **Boot > Distro Profiles**
2. Haz click en **"+ Add Custom Profile"**
3. Rellena los campos del perfil
4. Haz click en **"Create Profile"**

### Vía API

```bash
curl -H "Authorization: Bearer $TOKEN" -X POST http://localhost:8081/api/profiles/save \
  -H "Content-Type: application/json" \
  -d '{
    "profile_id": "my-distro",
    "display_name": "My Custom Distro",
    "family": "debian",
    "filename_patterns": ["mydistro", "my-distro"],
    "kernel_paths": ["/live/vmlinuz", "/boot/vmlinuz"],
    "initrd_paths": ["/live/initrd.img", "/boot/initrd"],
    "squashfs_paths": ["/live/filesystem.squashfs"],
    "default_boot_params": "boot=live initrd=initrd ip=dhcp",
    "boot_params_with_squashfs": "boot=live initrd=initrd fetch={{SQUASHFS}}",
    "auto_install_type": "preseed"
  }'
```

### Borrar perfiles custom

Solo se pueden borrar perfiles custom. Los perfiles built-in se restauran en la próxima actualización.

```bash
curl -H "Authorization: Bearer $TOKEN" -X DELETE "http://localhost:8081/api/profiles/delete?id=my-distro"
```

## Campos del perfil

| Campo | Requerido | Descripción |
|-------|----------|-------------|
| `profile_id` | Sí | Identificador único (p. ej., `ubuntu`, `my-distro`) |
| `display_name` | Sí | Nombre legible mostrado en la UI |
| `family` | No | Familia de distro (p. ej., `debian`, `arch`, `redhat`) — para agrupar |
| `filename_patterns` | Sí | Substrings a buscar en nombres de ISO (case-insensitive) |
| `kernel_paths` | No | Paths a probar para el kernel dentro del ISO (p. ej., `/casper/vmlinuz`) |
| `initrd_paths` | No | Paths a probar para el initrd dentro del ISO |
| `squashfs_paths` | No | Paths a probar para el filesystem root squashfs |
| `default_boot_params` | No | Parámetros de arranque del kernel por defecto (con soporte de placeholders) |
| `boot_params_with_squashfs` | No | Boot params alternativos usados cuando se detecta squashfs |
| `auto_install_type` | No | Formato de auto-instalación: `preseed`, `kickstart`, `autoinstall`, `autounattend` |
| `boot_method` | No | Override del método de arranque (p. ej., `wimboot` para Windows) |

## Placeholders

Los parámetros de arranque soportan estos placeholders, resueltos en el momento del arranque:

| Placeholder | Se resuelve a | Ejemplo |
|-------------|-------------|---------|
| `{{BASE_URL}}` | URL HTTP del servidor | `http://192.168.1.10:8080` |
| `{{CACHE_DIR}}` | Directorio de archivos extraídos | `ubuntu-24.04-server-amd64` |
| `{{FILENAME}}` | Filename del ISO (URL-encoded) | `ubuntu-24.04-server-amd64.iso` |
| `{{SQUASHFS}}` | URL completa al archivo squashfs | `http://192.168.1.10:8080/boot/ubuntu.../casper/filesystem.squashfs` |

### Ejemplo con placeholders

```
boot=live initrd=initrd fetch={{SQUASHFS}} ip=dhcp
```

Se resuelve a:
```
boot=live initrd=initrd fetch=http://192.168.1.10:8080/boot/debian-live-13/live/filesystem.squashfs ip=dhcp
```

## Ejemplos

### ISO live basada en Debian

```json
{
  "profile_id": "my-debian-live",
  "display_name": "My Debian Live Spin",
  "family": "debian",
  "filename_patterns": ["my-debian"],
  "kernel_paths": ["/live/vmlinuz"],
  "initrd_paths": ["/live/initrd.img"],
  "squashfs_paths": ["/live/filesystem.squashfs"],
  "default_boot_params": "initrd=initrd boot=live priority=critical",
  "boot_params_with_squashfs": "initrd=initrd boot=live priority=critical fetch={{SQUASHFS}}"
}
```

### Distro basada en Arch

```json
{
  "profile_id": "my-arch-spin",
  "display_name": "My Arch Spin",
  "family": "arch",
  "filename_patterns": ["myarch"],
  "kernel_paths": ["/arch/boot/x86_64/vmlinuz-linux", "/boot/vmlinuz-linux"],
  "initrd_paths": ["/arch/boot/x86_64/initramfs-linux.img", "/boot/initramfs-linux.img"],
  "squashfs_paths": ["/arch/x86_64/airootfs.sfs"],
  "default_boot_params": "archisobasedir=arch archiso_http_srv={{BASE_URL}}/boot/{{CACHE_DIR}}/iso/ ip=dhcp"
}
```

### Instalador basado en RHEL

```json
{
  "profile_id": "my-rhel-clone",
  "display_name": "My RHEL Clone",
  "family": "redhat",
  "filename_patterns": ["myrhel"],
  "kernel_paths": ["/images/pxeboot/vmlinuz"],
  "initrd_paths": ["/images/pxeboot/initrd.img"],
  "default_boot_params": "root=live:{{BASE_URL}}/isos/{{FILENAME}} rd.live.image inst.repo={{BASE_URL}}/boot/{{CACHE_DIR}}/iso/ rd.neednet=1 ip=dhcp",
  "auto_install_type": "kickstart"
}
```

## Solución de problemas

### ISO no detectado como la distro correcta

Comprueba si el filename del ISO coincide con algún patrón de perfil:

1. Ve a la pestaña **Distro Profiles**
2. Mira la columna "Filename Patterns"
3. Si ningún patrón coincide con el filename de tu ISO, crea un perfil custom

### Boot params incorrectos tras la extracción

1. Abre las **Properties** de la imagen
2. Haz click en **"Re-detect"** junto a Boot Parameters
3. O edita los boot params manualmente — soportan placeholders

### "Check for Updates" falló

La actualización descarga desde GitHub. Comprueba:
- El servidor tiene acceso a internet
- `raw.githubusercontent.com` no está bloqueado
- Inténtalo más tarde si GitHub está caído

### El perfil custom no coincide

Los perfiles custom tienen prioridad sobre los built-in. Asegúrate de que:
- Los `filename_patterns` contienen substrings que coinciden con tu filename ISO (case-insensitive)
- El profile ID es único
- El perfil se guardó correctamente

### Contribuir perfiles

Para añadir un perfil a la lista oficial para todos los usuarios:
1. Forkea el [repositorio de Bootimus](https://github.com/garybowers/bootimus)
2. Edita `distro-profiles.json` en la raíz del repo
3. Añade tu perfil al array `profiles`
4. Envía un pull request

Así todos los usuarios de Bootimus obtienen el nuevo perfil vía "Check for Updates".
