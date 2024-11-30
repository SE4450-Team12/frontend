import { app, BrowserWindow, ipcMain } from 'electron';
import path from 'path'; 
import isDev from 'electron-is-dev';
import database from 'better-sqlite3-multiple-ciphers';

app.on('ready', () => {
    const win = new BrowserWindow({
        width: 800,
        height: 600,
        webPreferences: {
            preload: path.join(app.getAppPath(), isDev ? '.' : '..', '/electron/preload.cjs')
        }
    });

    if (isDev) win.loadURL('http://localhost:3000');
    else win.loadFile(path.join(app.getAppPath(), '/dist-react/index.html'));
});

app.on('window-all-closed', () => {
    if (process.platform !== 'darwin') {
      app.quit();
    }
});
  
app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
});

ipcMain.handle("createDB", async (event, args) => {
    const db = new database(path.join(isDev ? app.getAppPath() : app.getPath("userData"), 'app.db'));
    db.pragma('journal_mode = WAL');
    // db.pragma(`rekey='${args[0]}'`);
    db.close();
});

ipcMain.handle("createDHTable", async (event, args) => {
    const db = new database(path.join(isDev ? app.getAppPath() : app.getPath("userData"), 'app.db'));
    // db.pragma(`key='${args[0]}'`);

    const query = `
        CREATE TABLE dh_keys (
            id INTEGER PRIMARY KEY,
            pubKey STRING NOT NULL,
            privKey STRING NOT NULL
        )`
    ;
    db.exec(query);
    db.close();

});

ipcMain.handle("insertDHKeys", async (event, args) => {
    const db = new database(path.join(isDev ? app.getAppPath() : app.getPath("userData"), 'app.db'));
    // db.pragma(`key='${args[0]}'`);
    const keys = args[0];

    for (let key in keys) {
        const insertData = db.prepare("INSERT INTO dh_keys (id, pub, priv) VALUES (?, ?, ?)").run(key.id, key.pubKey, key.privKey);
    }
    db.close();
});

ipcMain.handle("getDHKey", async (event, args) => {
    const db = new database(path.join(isDev ? app.getAppPath() : app.getPath("userData"), 'app.db'));
    // db.pragma(`key='${args[0]}'`);
    const row = db.prepare("SELECT * FROM dh_keys WHERE id = ?").get(args[0]);
    db.close();
    return row
});

ipcMain.handle("delDHKey", async (event, args) => {
    const db = new database(path.join(isDev ? app.getAppPath() : app.getPath("userData"), 'app.db'));
    // db.pragma(`key='${args[0]}'`);
    const delData = db.prepare("DELETE FROM dh_keys WHERE id = ?").run(args[0]);
    db.close();
});

ipcMain.handle("delAllDHKeys", async (event, args) => {
    const db = new database(path.join(isDev ? app.getAppPath() : app.getPath("userData"), 'app.db'));
    // db.pragma(`key='${args[0]}'`);
    const delData = db.prepare("DELETE FROM dh_keys").run();
    db.close();
});