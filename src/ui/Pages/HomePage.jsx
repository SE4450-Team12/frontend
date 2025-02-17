import { useEffect, useState } from "react";
import { Link } from "react-router";
import { io } from "socket.io-client";
import axios from "axios";
import { toast } from 'sonner';

import { generateDHKeys } from "@/Logic/WasmFunctions";

import { Chatroom } from "@/Components/Chatroom/Chatroom"
import { Sidebar } from "@/Components/Sidebar/Sidebar"
import { Button } from "@/Components/ui/button";
import { NewChat } from "@/Components/Sidebar/NewChat";
import { ChatNotifications } from "@/Components/Notifications/ChatNotifications";
import { SettingsModal } from "@/Components/Settings/SettingsModal";

import { ThemeProvider } from "styled-components";
import { lightTheme, darkTheme } from "@/Components/ui/themes"; 


export const HomePage = ({ loggedIn, username, userId, apiroot }) => {
    const [socket, setSocket] = useState();
    const [chatrooms, setChatrooms] = useState([]);
    const [currChatroom, setCurrChatroom] = useState();
    const [msgNotifs, setMsgNotifs] = useState({});
    const [addNewChatToggle, setAddNewChatToggle] = useState(false);
    const [newChatCreated, setNewChatCreated] = useState(false);
    const [showSettings, setShowSettings] = useState(false);
    const [darkMode, setDarkMode] = useState(false);
    const currentTheme = darkMode ? darkTheme : lightTheme;

    const toggleDarkMode = () => {
        setDarkMode((prevMode) => !prevMode);
      };

    useEffect(() => {
        async function setupSocket() {
            let tempSoc = await io("https://se4450.duckdns.org/", {
                    extraHeaders: {
                        Authorization: sessionStorage.getItem("JWT"),
                    }
                });
            setSocket(tempSoc);
        }
        if (loggedIn) {
            setupSocket()
        }
    }, [loggedIn]);

    useEffect(() => {
        if (socket != null) {
            socket.on('connect', async () => {
                let numKeys = 100
                const genDHKeys = await generateDHKeys(numKeys);
                const parsedKeys = JSON.parse(genDHKeys);
                if (parsedKeys.err != '') {
                    console.log(parsedKeys.err);
                    return toast.error('Gen Keys: Error generating DH Keys. Check console for error');
                }

                await window.electron.insertDHKeys(parsedKeys.keys);
                let keys = await window.electron.getKeys(numKeys);
                
                await sendDHKeysRequest(keys)
                await getChatroomsRequest(socket)
            });
        }
    }, [socket]);

    useEffect(() => {
        async function refreshSideBar() {
            if (newChatCreated) {
                await getChatroomsRequest(socket);
                setNewChatCreated(false);
            }
        };
        refreshSideBar();
    }, [newChatCreated]);

    const sendDHKeysRequest = async (keys) => {
        let response;
        try {
            response = await axios.put(`${apiroot}/user/dh_keys`, keys, {
                headers: {
                    Authorization: sessionStorage.getItem("JWT"),
                },
            });
        } catch(err) {
            console.log(err);
            return toast.error("Gen Keys: Failed to send keys to server. Check console for error");
        }
    };

    const getChatroomsRequest = async (soc) => {
        let response;
        try {
            response = await axios.get(`${apiroot}/chatroom`, {
                headers: {
                    Authorization: sessionStorage.getItem("JWT"),
                },
            });
        } catch(err) {
            toast.error("Error getting chatrooms. Check Console");
            console.log(err);
            return;
        }

        setChatrooms(response.data);
        for (let room of response.data) {
            soc.emit("joinRoom", { "chatroomId": room._id });
        }
    }

    const logout = () => {
        socket.disconnect();
    }

    const test = () => {
        console.log(addNewChat);
    }

    return (
        <ThemeProvider theme={currentTheme}>
          <div className={`flex flex-row h-screen relative overflow-hidden ${darkMode ? 'dark' : ''}`}>
            <NewChat
              isOpen={addNewChatToggle}
              toggle={setAddNewChatToggle}
              apiroot={apiroot}
              setNewChatCreated={setNewChatCreated}
              setCurrChatroom={setCurrChatroom}
            />
            <Sidebar
              username={username}
              chatrooms={chatrooms}
              currChatroom={currChatroom}
              setCurrChatroom={(room) => {
                setMsgNotifs((prev) => ({ ...prev, [room._id]: false }));
                setCurrChatroom(room);
              }}
              msgNotifs={msgNotifs}
              setAddNewChat={setAddNewChatToggle}
              isNewChatOpen={addNewChatToggle}  
            />
            <Chatroom
              chatroom={currChatroom}
              userId={userId}
              socket={socket}
              setMsgNotifs={setMsgNotifs}
              apiroot={apiroot}
            />
            {socket && (
              <ChatNotifications
                socket={socket}
                userId={userId}
                currentChatroomId={currChatroom ? currChatroom._id : undefined}
                chatrooms={chatrooms}
                setMsgNotifs={setMsgNotifs}
              />
            )}
            <div className="fixed top-[10px] right-[10px] flex gap-2">
              <Button
                className="py-[1px] px-[10px] bg-blue-600 hover:bg-blue-700"
                onClick={() => setShowSettings(true)}
              >
                Settings
              </Button>
              <Button
                className="py-[1px] px-[10px] bg-red-600 hover:bg-red-700"
                onClick={() => socket.disconnect()}
              >
                <Link to="/">Log out</Link>
              </Button>
            </div>
            {showSettings && (
              <SettingsModal
                onClose={() => setShowSettings(false)}
                darkMode={darkMode}
                toggleDarkMode={toggleDarkMode}
                apiroot={apiroot}
              />
            )}
          </div>
        </ThemeProvider>
      );
      
    };