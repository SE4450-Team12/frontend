import { useEffect, useState, useRef } from 'react';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import axios from 'axios';

import '@/Components/Chatroom/Chatroom.css'
import { encryptMsg, decryptMsg } from '@/Logic/WasmFunctions';

import { Button } from '@/Components/ui/button';
import { 
    Form,
    FormControl,
    FormField,
    FormItem
} from '@/Components/ui/form';
import { Input } from '@/Components/ui/input';
import { Message } from '@/Components/Message/Message';


export const formSchema = z.object({
    msg: z.string().min(1),
});

export const Chatroom = ({ chatroom, userId, socket, setMsgNotifs, apiroot, newChatMembers }) => {
    const [messages, setMessages] = useState([]);
    const [chatMember, setChatMember] = useState("");
    const outMsgKeys = useRef({});

    const msgInputRef = useRef(null);
    const chatBottom = useRef(null);
    

    const form = useForm({
        resolver: zodResolver(formSchema),
        defaultValues: {
            msg: "",
        },
    });

    useEffect(() => {
        async function socketNewMsg() {
            socket.on('newMessage', await handleMsgIn);
        }
        if (socket != null && chatroom != null) {
            getMessages();
            socketNewMsg();
            console.log(chatroom);
        }
    }, [socket, chatroom]);

    const addMessage = async (msg) => {
        await window.electron.insertMsg(msg);
        let parsedMsg = await window.electron.getMsg(msg._id);

        if (msg.chatroom === chatroom._id) {
            setMessages((prevMsgs) => [...prevMsgs, parsedMsg]);
            setTimeout(() => {
                chatBottom.current.scrollIntoView({ behaviour: 'smooth' });
            }, 10);
        }
    };

    const handleMsgIn = async (data) => {
        if (data.sender === userId) {
            let hash = await window.electron.sha256(data.message.content + data.message.pubKey + data.message.timestamp);
            if (outMsgKeys.current[hash]) {
                let res = await decryptMsg(data.message.content, data.message.timestamp, outMsgKeys.current[hash], "");
                if (res["error"] != "") {
                    console.log(res["error"]);
                    return toast.error("Msg In: Failed to decrypt msg. Check console for error");
                }
        
                delete outMsgKeys.current[hash];
                data.message.content = res["plainText"];
                await addMessage(data);
            }
            return;
        }
        
        let myKey = await electron.getDHKey(parseInt(data.message.privKeyId));

        let res = await decryptMsg(data.message.content, data.message.timestamp, data.message.pubKey, myKey.privKey);
        if (res["error"] != "") {
            console.log(res["error"]);
            return toast.error("Msg In: Failed to decrypt msg. Check console for error");
        }
        data.message.content = res["plainText"];
        if (data.chatroom != chatroom._id) {
            setMsgNotifs((prevNotifs) => ({ ...prevNotifs, [data.chatroom]: true}));
            return;
        }
        await addMessage(data)
    };

    const getMessages = async () => {
        if (chatroom == null) {
            setMessages([]);
            return;
        }

        for (let mem of chatroom.members) {
            if (mem != userId) {
                setChatMember(mem);
                break;
            }
        }

        let response
        try {
            response = await axios.get(`${apiroot}/message/${chatroom._id}`, {
                headers: {
                    Authorization: sessionStorage.getItem("JWT"),
                }   
            });
        } catch(err) {
            toast.error("Error getting messages. Check Console");
            console.log(err);
            setMessages([]);
            return;
        }

        const currentChat = response.data;
        let storedChat = await window.electron.getMsgs(chatroom._id);

        if (currentChat.length > storedChat.length) {
            for (let i = storedChat.length; i < currentChat.length; i++) {
              let myKey = await window.electron.getDHKey(parseInt(currentChat[i].message.privKeyId));
    
              let res = await decryptMsg(currentChat[i].message.content, currentChat[i].message.timestamp, currentChat[i].message.pubKey, myKey.privKey);
              if (res["error"] != "") {
                console.log(res["error"]);
                return toast.error("Msg In: Failed to decrypt msg. Check console for error");
              }
              currentChat[i].message.content = res["plainText"];
              await window.electron.insertMsg(currentChat[i]);
              currentChat[i] = {};
            }
          }

        let finalChat = await window.electron.getMsgs(chatroom._id);
        setMessages(finalChat);
        setTimeout(() => {
            chatBottom.current.scrollIntoView({ behaviour: 'smooth' });
        }, 10);

    };


    const sendMessage = async (values) => {
        let response;

        try {
            response = await axios.delete(`${apiroot}/user/dh_keys/${chatMember}`, {
                headers: {
                    Authorization: sessionStorage.getItem("JWT"),
                }
        });
        } catch (err) {
            toast.error("Sending Msg: Failed to get other user's key. Check console for error");
            console.log(err);
            return
        }

        const otherPubDH = response.data.popped_key.pubKey;
        const otherPubDHId = response.data.popped_key.id; 
        const timestamp = new Date();
        
        let encData = await encryptMsg(otherPubDH, values.msg, timestamp.toJSON());
        if (encData["error"] != "") {
            toast.error("Sending Msg: Failed to encrypt message. Check console for error");
            console.log(encData["error"]);
            return
        }
    
        let payload = {
            "content": encData["cipherText"],
            "pubKey": encData["pubKey"],
            "privKeyId": otherPubDHId.toString(),
            "timestamp": timestamp.toJSON(),
        };
    
        let hash = await window.electron.sha256(payload.content + payload.pubKey + payload.timestamp);
        let val = {};
        val[`${hash}`] = encData["masterSec"];
        outMsgKeys.current = {...outMsgKeys.current, ...val};

        const pendingMessage = {
            _id: "pending_" + new Date().getTime(),
            content: values.msg,
            send: userId,
            pending: true,
            timestamp: timestamp.toJSON(),
            hash: hash
        };

        setMessages((prevMessages) => [...prevMessages, pendingMessage])

        socket.emit("chatroomMessage", {
            "chatroomId": chatroom._id,
            "message": payload
        },
            (ackReponse) => {
                if (ackReponse && ackReponse.status === "ok" && ackResponse.messageData) {
                    setMessages((prevMessages) =>
                        prevMessages.map((msg) => {
                            if (msg._id === pendingMessage._id) {
                                return {...ackResponse.messageData, pending: false};
                            }
                            return msg;
                        })
                    );
                }   else {
                    setMessages((prevMessages) =>
                        prevMessages.filter((msg) => msg._id !== pendingMessage._id)
                    );
                    toast.error("Message sending failed. Please try again.");
                    }
                }
            ); 

        form.reset();
        msgInputRef.current.focus();
    };

    const test = () => {
        console.log(chatroom);
    };

    return(
        <div className="chatroom">
            <h1 className="text-center text-4xl font-bold mb-10 text-green-600">Chatroom</h1>
            <div className="flex-1 overflow-y-scroll p-5 box-border">
                {messages == null || messages.length == 0 ? 'No chats to show' : messages.map((msg) => (
                    <Message content={msg.content} isSender={msg.sender == userId} key={msg.mongoId} pending={msg.pending}/>
                ))}
                <div ref={chatBottom}/>
            </div>
            {chatroom == null && newChatMembers?.length == 0 ? '' : 
                <Form {...form}>
                    <form onSubmit={form.handleSubmit(sendMessage)} className="flex p-5 bg-white border-solid border-t border-gray-300">
                        <FormField
                            control={form.control}
                            name="msg"
                            render={({ field }) => (
                                <FormItem className="w-full">
                                    <FormControl>
                                        <Input {...field} className="flex-1 text-base" ref={msgInputRef}/>
                                    </FormControl>
                                </FormItem>
                            )}
                        />
                        <Button type="submit" className="text-base">Send</Button>
                        {/* <Button onClick={test}>Test</Button> */}
                    </form>
                </Form>
            }
        </div>
    );
};