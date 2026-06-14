# AGENTS.md — Mobile (Android)

Android app — Kotlin 2.2.10, Jetpack Compose, Material3, AGP 9.2.1.

Read `mobile/docs/` for topic-specific patterns before writing any Android code.

Claude Code users: see `CLAUDE.md` for the feature development workflow and subagent definitions.

---

## Setup

Prerequisites: Android Studio Meerkat (2024.3) or newer, JDK 17+, Android SDK with API 36.

```bash
# Build debug APK
cd mobile && ./gradlew assembleDebug

# Install on connected device or emulator
cd mobile && ./gradlew installDebug
```

Open the project in Android Studio by pointing it at the `mobile/` directory.

---

## Commands

```bash
cd mobile && ./gradlew assembleDebug        # compile debug build
cd mobile && ./gradlew lint                 # run Android lint
cd mobile && ./gradlew test                 # unit tests (no device needed)
cd mobile && ./gradlew connectedAndroidTest # instrumented tests (device/emulator required)
cd mobile && ./gradlew clean                # clean build outputs
```

On Windows, use `.\gradlew.bat` instead of `./gradlew` if running outside Git Bash.

---

## Project structure

```
mobile/
  app/
    build.gradle.kts                        # app dependencies and build config
    src/
      main/
        AndroidManifest.xml
        java/com/company/template/
          MainActivity.kt                   # single entry point, Compose root
          ui/
            theme/
              Color.kt                      # color palette constants
              Theme.kt                      # TemplateTheme composable
              Type.kt                       # Typography object
      test/                                 # unit tests (JVM only)
      androidTest/                          # instrumented tests (device/emulator)
  gradle/
    libs.versions.toml                      # version catalog — all versions declared here
  build.gradle.kts                          # root build config
  settings.gradle.kts                       # module declarations
```

---

## Adding a new screen

1. Create `ui/<FeatureName>Screen.kt` in the main source set.
2. Define a `@Composable` function named `<FeatureName>Screen(modifier: Modifier = Modifier)`.
3. Add a `@Preview(showBackground = true)` at the bottom.
4. Wire it into the navigation or call it from `MainActivity` during development.

## Adding a new dependency

1. Declare the version in `gradle/libs.versions.toml` under `[versions]`.
2. Add the library alias under `[libraries]`.
3. Reference it in `app/build.gradle.kts` via `libs.<alias>`.
4. Compose BOM libraries do not need a version — the BOM manages them.

---

## Key conventions

- **Single Activity** — all screens are Composables, no Fragment stack.
- **No logic in Composables** — state and business logic belong in ViewModels or the calling composable; keep UI functions pure.
- **Material3 only** — use `androidx.compose.material3`, not the older `material` package.
- **Theme tokens** — use `MaterialTheme.colorScheme.*` and `MaterialTheme.typography.*`; never hardcode colors or font sizes in screen files.
- **Version catalog** — all dependency versions must be declared in `libs.versions.toml`, not as raw strings in `build.gradle.kts`.
- **Modifier as last parameter** — always accept `modifier: Modifier = Modifier` as the last defaulted parameter in public Composable functions.

---

## Testing

**TDD is required.** Write failing tests first, then implement.

- **Unit tests** (`src/test/`): JUnit 4, JVM only. Use for pure Kotlin logic and ViewModels. Run with `./gradlew test`.
- **Instrumented tests** (`src/androidTest/`): JUnit 4 + Compose test rules. Use for Composables and Activity-level tests. Run with `./gradlew connectedAndroidTest` — requires a running emulator or connected device.
- No mocking of the Android framework — use `InstrumentationRegistry` for context.
- Every new public `@Composable` function must have a corresponding instrumented test.

```kotlin
// Composable instrumented test pattern
@RunWith(AndroidJUnit4::class)
class GreetingTest {
    @get:Rule val composeTestRule = createComposeRule()

    @Test
    fun greeting_displaysName() {
        composeTestRule.setContent {
            TemplateTheme { Greeting(name = "World") }
        }
        composeTestRule.onNodeWithText("Hello World!").assertIsDisplayed()
    }
}
```

See [`docs/testing.md`](docs/testing.md) for the full pattern and conventions.

---

## Security

- No hardcoded API keys, tokens, or secrets anywhere in source files or `build.gradle.kts`.
- Store secrets in `local.properties` (gitignored) and read via `BuildConfig` fields in `build.gradle.kts`.
- All network calls must go over HTTPS.
