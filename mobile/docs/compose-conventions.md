---
topic: Jetpack Compose UI conventions
last_verified: 2026-06-14
sources:
  - app/src/main/java/com/company/template/MainActivity.kt
  - app/src/main/java/com/company/template/ui/theme/Theme.kt
  - app/src/main/java/com/company/template/ui/theme/Color.kt
  - app/src/main/java/com/company/template/ui/theme/Type.kt
---

# Jetpack Compose UI conventions

## Theme

All UI is wrapped in `TemplateTheme` exactly once — at the `setContent` call in `MainActivity`. Do not call `TemplateTheme` inside individual screens or components.

```kotlin
// MainActivity.kt — only place TemplateTheme is applied
setContent {
    TemplateTheme {
        Scaffold(modifier = Modifier.fillMaxSize()) { innerPadding ->
            MyScreen(modifier = Modifier.padding(innerPadding))
        }
    }
}
```

`TemplateTheme` accepts:
- `darkTheme: Boolean` — defaults to `isSystemInDarkTheme()`
- `dynamicColor: Boolean` — defaults to `true`; uses Material You wallpaper colors on Android 12+ (SDK ≥ 31); falls back to the static `Purple/Pink` palette on older devices

## Color palette

Static fallback colors are defined in `ui/theme/Color.kt`:

```kotlin
val Purple80 = Color(0xFFD0BCFF)   // dark theme primary
val PurpleGrey80 = Color(0xFFCCC2DC)
val Pink80 = Color(0xFFEFB8C8)

val Purple40 = Color(0xFF6650A4)   // light theme primary
val PurpleGrey40 = Color(0xFF625B71)
val Pink40 = Color(0xFF7D5260)
```

In screens and components, never reference these constants directly. Access colors via `MaterialTheme.colorScheme.*` so dynamic color and dark mode are respected:

```kotlin
Text(
    text = "Hello",
    color = MaterialTheme.colorScheme.onBackground
)
```

## Typography

`Typography` is defined in `ui/theme/Type.kt` with `bodyLarge` overridden (16sp, Normal weight). Access via `MaterialTheme.typography.*`:

```kotlin
Text(
    text = "Label",
    style = MaterialTheme.typography.bodyLarge
)
```

## Composable function conventions

### Signature
- Accept `modifier: Modifier = Modifier` as the last defaulted parameter on every public `@Composable`.
- Hoist state — `@Composable` functions must be stateless; pass values and lambdas in.

```kotlin
@Composable
fun Greeting(name: String, modifier: Modifier = Modifier) {
    Text(
        text = "Hello $name!",
        modifier = modifier
    )
}
```

### Screen vs component
- **Screens** — top-level Composables called from MainActivity or a nav host. File name: `<Feature>Screen.kt`.
- **Components** — reusable pieces. Place in `ui/components/`. Accept a `modifier` parameter; use Material3 primitives.

### Previews
Every public Composable must have a `@Preview`:

```kotlin
@Preview(showBackground = true)
@Composable
fun GreetingPreview() {
    TemplateTheme {
        Greeting("Android")
    }
}
```

Always wrap previews in `TemplateTheme` so colors and typography resolve correctly.

## Scaffold

Use `Scaffold` as the root layout for screens that need a top bar, bottom bar, or FAB:

```kotlin
Scaffold(modifier = Modifier.fillMaxSize()) { innerPadding ->
    ContentComposable(modifier = Modifier.padding(innerPadding))
}
```

Pass `innerPadding` down to the content composable via `Modifier.padding(innerPadding)` — never ignore it.

## Material3

Import only from `androidx.compose.material3`:
```kotlin
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
```

Never import from `androidx.compose.material` (M2) — both are in the dependency tree but only M3 is used here.
