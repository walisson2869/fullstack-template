---
topic: Activity and Compose architecture
last_verified: 2026-06-14
sources:
  - app/src/main/java/com/company/template/MainActivity.kt
  - app/build.gradle.kts
  - gradle/libs.versions.toml
---

# Activity and Compose architecture

## Entry point

`MainActivity` is the single Activity for the app. It extends `ComponentActivity` and sets up the Compose UI tree in `onCreate`:

```kotlin
class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            TemplateTheme {
                Scaffold(modifier = Modifier.fillMaxSize()) { innerPadding ->
                    Greeting(
                        name = "Android",
                        modifier = Modifier.padding(innerPadding)
                    )
                }
            }
        }
    }
}
```

Key calls:
- `enableEdgeToEdge()` — called before `setContent`; allows content to draw behind system bars.
- `setContent { }` — replaces the XML layout system; the lambda is the Compose UI root.
- `TemplateTheme { }` — applied once here; all screens inherit the theme automatically.

## Single-Activity pattern

There is no `Fragment` stack. All navigation between screens happens inside the Compose composition (via Navigation Compose when added). Do not create additional Activities or Fragments.

## Lifecycle

`ComponentActivity` integrates with Jetpack Lifecycle. When adding state or coroutines:
- Use `viewModel()` (from `lifecycle-viewmodel-compose`) to scope ViewModels to the Activity or a nav destination.
- Collect `StateFlow` / `Flow` from ViewModels using `collectAsStateWithLifecycle()` (from `lifecycle-runtime-compose`) — not `collectAsState()`, which does not respect lifecycle.

## Dependency versions

All versions are declared in `gradle/libs.versions.toml`. Current versions:

| Library | Version |
|---|---|
| Kotlin | 2.2.10 |
| AGP | 9.2.1 |
| Compose BOM | 2026.02.01 |
| `androidx.core:core-ktx` | 1.10.1 |
| `androidx.lifecycle:lifecycle-runtime-ktx` | 2.6.1 |
| `androidx.activity:activity-compose` | 1.8.0 |

Compose library versions (ui, material3, etc.) are managed by the BOM — do not pin them individually.

## Build configuration

- `compileSdk` 36, `minSdk` 24, `targetSdk` 36
- Source and target compatibility: Java 11
- `buildFeatures { compose = true }` — enables the Compose compiler
- Kotlin plugin: `org.jetbrains.kotlin.plugin.compose` (separate from the language plugin; required for Compose)

## Adding a ViewModel (when needed)

1. Add `lifecycle-viewmodel-compose` to `libs.versions.toml` and `app/build.gradle.kts`.
2. Create `ui/<Feature>ViewModel.kt` extending `ViewModel`.
3. Expose UI state as `StateFlow<UiState>`.
4. Inject into the Composable via `viewModel()`:

```kotlin
@Composable
fun FeatureScreen(viewModel: FeatureViewModel = viewModel()) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    // render uiState
}
```
