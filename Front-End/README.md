# Getting Started with Create React App

This project was bootstrapped with [Create React App](https://github.com/facebook/create-react-app).

## Optional local Manga Hook API

[Manga Hook](https://github.com/kiraaziz/mangahook-api) is an external,
scraper-based manga API that we may use as a secondary provider. TaleHaven
currently uses MangaDex only, so starting Manga Hook does not route any
TaleHaven traffic to it yet. This section is only for running and checking the
server locally ahead of that integration.

1. Install a current Node.js LTS release, which includes `npm`.
2. Clone Manga Hook outside this repository and open its server directory:

   ```powershell
   git clone https://github.com/kiraaziz/mangahook-api.git C:\src\mangahook-api
   Set-Location C:\src\mangahook-api\server
   ```

3. Install its dependencies:

   ```powershell
   npm install
   ```

4. Ensure `server/.env` contains a port that does not conflict with TaleHaven.
   The recommended local value is `PORT=3002` (`3000` is the frontend and
   `3001` is the TaleHaven backend).
5. Start the server:

   ```powershell
   npm run start
   ```

6. With the server running, make a quick local request in another PowerShell
   window:

   ```powershell
   Invoke-RestMethod http://localhost:3002/api/mangaList
   ```

Manga Hook scrapes an upstream website, so a successful local startup does not
guarantee every request will return data. Do not expose this development server
directly to the public internet. The eventual TaleHaven integration should call
it from the backend, apply timeouts and rate limits, and fall back to MangaDex
when appropriate.

## Available Scripts

In the project directory, you can run:

### `npm start`

Runs the app in the development mode.\
Open [http://localhost:3000](http://localhost:3000) to view it in the browser.

The page will reload if you make edits.\
You will also see any lint errors in the console.

### `npm test`

Launches the test runner in the interactive watch mode.\
See the section about [running tests](https://facebook.github.io/create-react-app/docs/running-tests) for more information.

### `npm run build`

Builds the app for production to the `build` folder.\
It correctly bundles React in production mode and optimizes the build for the best performance.

The build is minified and the filenames include the hashes.\
Your app is ready to be deployed!

See the section about [deployment](https://facebook.github.io/create-react-app/docs/deployment) for more information.

### `npm run eject`

**Note: this is a one-way operation. Once you `eject`, you can’t go back!**

If you aren’t satisfied with the build tool and configuration choices, you can `eject` at any time. This command will remove the single build dependency from your project.

Instead, it will copy all the configuration files and the transitive dependencies (webpack, Babel, ESLint, etc) right into your project so you have full control over them. All of the commands except `eject` will still work, but they will point to the copied scripts so you can tweak them. At this point you’re on your own.

You don’t have to ever use `eject`. The curated feature set is suitable for small and middle deployments, and you shouldn’t feel obligated to use this feature. However we understand that this tool wouldn’t be useful if you couldn’t customize it when you are ready for it.

## Learn More

You can learn more in the [Create React App documentation](https://facebook.github.io/create-react-app/docs/getting-started).

To learn React, check out the [React documentation](https://reactjs.org/).
