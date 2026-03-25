package com.legoparser

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import com.legoparser.ui.navigation.AppNavigation
import com.legoparser.ui.theme.LegoParserTheme

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            LegoParserTheme {
                AppNavigation()
            }
        }
    }
}
