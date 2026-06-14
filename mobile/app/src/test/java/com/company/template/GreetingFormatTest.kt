package com.company.template

import org.junit.Test
import org.junit.Assert.*

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
