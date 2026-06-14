---
topic: Testing patterns
last_verified: 2026-06-15
sources:
  - app/src/test/java/com/company/template/GreetingFormatTest.kt
  - app/src/androidTest/java/com/company/template/GreetingTest.kt
  - app/build.gradle.kts
  - gradle/libs.versions.toml
---

# Testing patterns

## TDD mandate

Write a failing test before writing the implementation. Every new public `@Composable` must have a corresponding instrumented test in `src/androidTest/`.

## Two test source sets

| Source set | Path | Runs on | Framework |
|---|---|---|---|
| Unit tests | `src/test/` | JVM (host machine) | JUnit 4 |
| Instrumented tests | `src/androidTest/` | Device or emulator | JUnit 4 + Espresso + Compose test |

## Unit tests (`src/test/`)

Run with `./gradlew test`. No Android framework available — test pure Kotlin logic here.

`GreetingFormatTest.kt` is the canonical example (replaces the scaffold `ExampleUnitTest.kt`):

```kotlin
class GreetingFormatTest {

    @Test
    fun greeting_text_contains_name() {
        val name = "World"
        val expected = "Hello $name!"
        assertEquals("Hello World!", expected)
    }

    @Test
    fun greeting_text_with_empty_name() {
        val name = ""
        val expected = "Hello $name!"
        assertEquals("Hello !", expected)
    }
}
```

Place unit tests in the same package as the code under test. No mocking of Android framework classes — if a class requires Android context, it belongs in an instrumented test.

## Instrumented tests (`src/androidTest/`)

Run with `./gradlew connectedAndroidTest`. Requires a running emulator or physically connected device.

`GreetingTest.kt` is the canonical Compose UI test (replaces the scaffold `ExampleInstrumentedTest.kt`):

```kotlin
@RunWith(AndroidJUnit4::class)
class GreetingTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun greeting_displaysName() {
        composeTestRule.setContent {
            TemplateTheme {
                Greeting(name = "World")
            }
        }
        composeTestRule.onNodeWithText("Hello World!").assertIsDisplayed()
    }

    @Test
    fun greeting_displaysAndroid() {
        composeTestRule.setContent {
            TemplateTheme {
                Greeting(name = "Android")
            }
        }
        composeTestRule.onNodeWithText("Hello Android!").assertIsDisplayed()
    }
}
```

Use `createComposeRule()` (no Activity) for component tests. Use `createAndroidComposeRule<MainActivity>()` for end-to-end tests that require the full Activity.

## New Composable checklist

When adding a new public `@Composable`:

1. Create the composable in the appropriate file under `app/src/main/`.
2. Add an instrumented test class in `src/androidTest/` using `createComposeRule()`.
3. Test at minimum: the composable renders its primary content given representative inputs.
4. Run `./gradlew connectedAndroidTest` to confirm the test passes on a device/emulator.

## Dependencies

Test dependencies in `build.gradle.kts`:
```kotlin
testImplementation(libs.junit)                                          // unit tests
androidTestImplementation(platform(libs.androidx.compose.bom))
androidTestImplementation(libs.androidx.compose.ui.test.junit4)        // Compose UI tests
androidTestImplementation(libs.androidx.espresso.core)                  // Espresso
androidTestImplementation(libs.androidx.junit)                          // AndroidJUnit4 runner
debugImplementation(libs.androidx.compose.ui.test.manifest)             // required for Compose tests
```

## Running tests

```bash
# Unit tests — no device needed
cd mobile && ./gradlew test

# Instrumented tests — device/emulator required
cd mobile && ./gradlew connectedAndroidTest

# Both
cd mobile && ./gradlew check
```
