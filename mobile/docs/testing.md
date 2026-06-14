---
topic: Testing patterns
last_verified: 2026-06-14
sources:
  - app/src/test/java/com/company/template/ExampleUnitTest.kt
  - app/src/androidTest/java/com/company/template/ExampleInstrumentedTest.kt
  - app/build.gradle.kts
  - gradle/libs.versions.toml
---

# Testing patterns

## Two test source sets

| Source set | Path | Runs on | Framework |
|---|---|---|---|
| Unit tests | `src/test/` | JVM (host machine) | JUnit 4 |
| Instrumented tests | `src/androidTest/` | Device or emulator | JUnit 4 + Espresso + Compose test |

## Unit tests (`src/test/`)

Run with `./gradlew test`. No Android framework available — test pure Kotlin logic here.

```kotlin
class ExampleUnitTest {
    @Test
    fun addition_isCorrect() {
        assertEquals(4, 2 + 2)
    }
}
```

Place unit tests in the same package as the code under test. No mocking of Android framework classes — if a class requires Android context, it belongs in an instrumented test.

## Instrumented tests (`src/androidTest/`)

Run with `./gradlew connectedAndroidTest`. Requires a running emulator or physically connected device.

```kotlin
@RunWith(AndroidJUnit4::class)
class ExampleInstrumentedTest {
    @Test
    fun useAppContext() {
        val appContext = InstrumentationRegistry.getInstrumentation().targetContext
        assertEquals("com.company.template", appContext.packageName)
    }
}
```

Use `InstrumentationRegistry.getInstrumentation().targetContext` for the app's `Context`.

## Compose UI tests

For UI tests, use the Compose testing library (`androidx.compose.ui:ui-test-junit4`), already included in `build.gradle.kts`. Add a `ComposeTestRule` and test composables in isolation:

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
}
```

Use `createComposeRule()` (no Activity) for component tests. Use `createAndroidComposeRule<MainActivity>()` for end-to-end tests that require the full Activity.

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
