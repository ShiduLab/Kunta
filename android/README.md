# Kunta Android

Progetto Android autonomo.

- Java 17
- Android Gradle Plugin 8.7.3
- compile/target SDK 35
- min SDK 23
- UI Kunta incorporata in `app/src/main/assets/`
- nessun collegamento alla PWA online
- nessun `setup-android@v3`
- nessuna dipendenza da una cartella esterna al progetto Android

La GitHub Action nella root usa Gradle 8.9 fornito direttamente da `gradle/actions/setup-gradle`.
Non richiede `gradlew` nel repository.
