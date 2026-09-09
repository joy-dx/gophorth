import {useEffect, useMemo, useState} from 'react';
import './App.css';
import {releaserdto, updaterdto} from "@wailsmodels";
import {CheckForUpdate, StartUpdate, UpdaterStateGet} from "@wailsapi/UpdaterInterface";
import {UseLogStore} from "@lib";

const LogEntries = () => {
    const logEntries = UseLogStore((s) => s.logs["system"])
    return (
        <ul className="text-left">
            {logEntries.map((msg) => (
                <li
                    className= "whitespace-pre-wrap wrap-break-word rounded text-sm text-gray-800"
                    key={msg}
                >
                    {msg}
                </li>))}
        </ul>
    )
}

function App() {
    const [updaterState, setUpdaterState] = useState<updaterdto.UpdaterState>();
    const [responseText, setResponseText] = useState("");
    const [isUpdateAvailable, setIsUpdateAvailable] = useState<boolean>(false);
    const [releaseAsset, setReleaseAsset] = useState<releaserdto.ReleaseAsset>();
    const [isBusy, setIsBusy] = useState<boolean>(false);

    useEffect(() => {
        UpdaterStateGet().then(setUpdaterState, setResponseText)
    }, []);

    function onUpdateCheckClick() {
        setIsBusy(true)
        CheckForUpdate().then((releaseAsset: releaserdto.ReleaseAsset) => {
            if(releaseAsset.version !== updaterState?.updater_version) {
                setIsUpdateAvailable(true);
            }
            setReleaseAsset(releaseAsset);
        }, setResponseText).finally(() => {
            setIsBusy(false)
        });
    }

    function onUpdateStartClick() {
        setIsBusy(true)
        StartUpdate().then(() => {
           setResponseText("app will automatically shutdown and update in 3 second")
        }, setResponseText).finally(() => {
            setIsBusy(false)
        });
    }

    const updateLog = updaterState?.updater_log

    return (
        <main className="mx-auto max-w-6xl p-6">
            <div
                className="grid grid-cols-1 gap-6 md:grid-cols-2"
            >
                <section
                    className="rounded-lg border border-gray-200 bg-white p-6 shadow-sm"
                >
                    <h1 className="text-2xl font-semibold tracking-tight text-black">
                        App Title
                    </h1>

                    <div className="mt-2 text-sm text-gray-600">
                        Version: <span id="appVersion" className="font-mono">{updaterState?.updater_version}</span>
                    </div>

                    <div className="mt-6">
                        <div className="flex gap-x-2 justify-center">
                            <button
                                onClick={onUpdateCheckClick}
                                disabled={isBusy}
                                id="checkUpdatesBtn"
                                className="inline-flex items-center rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:cursor-not-allowed disabled:opacity-60"
                            >
                                Check for updates
                            </button>
                            {isUpdateAvailable && <button
                                disabled={isBusy}
                                className="inline-flex items-center rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:cursor-not-allowed disabled:opacity-60"
                                onClick={onUpdateStartClick}>Start update</button>}
                        </div>

                        {isUpdateAvailable && releaseAsset && (<>
                        <p
                            id="updateStatus"
                            className="mt-3 text-sm text-gray-700"
                        >
                            Found version {releaseAsset.version} to be downloaded from {releaseAsset.download_url}
                        </p>
                        <p
                            className="mt-3 text-sm text-gray-700"
                        >Once you initiate the update, after preparation the app will automatically restart and update and present the update log.</p>
                        <p
                            className="mt-3 text-sm text-gray-700"
                        >
                            There is a PostInstallCleanup function exposed from the updater interface for cleaning up
                        </p>
                        </>
                        )}
                        {responseText && <p className="mt-3 text-sm text-gray-700">{responseText}</p>}
                    </div>
                    {updateLog && <div className="mt-6 text-left text-black">
                        <p className="whitespace-pre-wrap">{updateLog}</p>
                    </div>}
                </section>

                <section
                    className="rounded-lg border border-gray-200 bg-white p-6 shadow-sm"
                >
                    <div className="mb-3 flex flex-col">
                        <h2 className="text-lg font-semibold text-black">Logs</h2>
                        <LogEntries/>
                    </div>
                </section>
            </div>
        </main>
)
}

export default App
