import { useEffect, useRef, useState } from 'react';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import axios from 'axios';

import { IoIosArrowRoundBack } from "react-icons/io";
import { IoMdCheckmark } from "react-icons/io";
import { Wrapper } from "@/Components/Sidebar/NewChat.styles";
import { 
    Form,
    FormControl,
    FormField,
    FormItem
} from '@/Components/ui/form';
import { Input } from '@/Components/ui/input';
import { toast } from 'sonner';
import { Button } from '@/Components/ui/button';

export const NewChat = ({ isOpen, toggle, apiroot, setNewChatCreated, setCurrChatroom }) => {
    const searchInputRef = useRef(null);
    const [usernames, setUsernames] = useState([]);
    const [selectedUser, setSelectedUser] = useState({});

    const formSchema = z.object({
        usernameSearch: z.string().min(1),
    });
    const form = useForm({
        resolver: zodResolver(formSchema),
        defaultValues: {
            usernameSearch: "",
        },
    });

    useEffect(() => {
        if (isOpen) {
            form.reset();
            searchInputRef.current.focus();
            setUsernames([]);
            setSelectedUser({});
        }
    }, [form, isOpen]);

    const searchUserReq = async (values) => {
        let response;
        try {
            response = await axios.get(`${apiroot}/user/name/${values.usernameSearch}`, {
                headers: {
                    Authorization: sessionStorage.getItem("JWT"),
                }
            });
        } catch(err) {
            if (err.status == 404) {
                setUsernames([]);
                return toast.error("User not found");
            }
        }
        setUsernames(response.data);
    };

    const startChatReq = async () => {
        if (Object.keys(selectedUser).length === 0) 
          return;
      
        let response;
        try {
          response = await axios.post(
            `${apiroot}/chatroom`,
            {
              members: [selectedUser._id] 
            },
            {
              headers: {
                Authorization: sessionStorage.getItem("JWT"),
              }
            }
          );
        } catch (err) {
          console.log(err);
          return toast.error("Chatroom could not be created.");
        }
        console.log(response);
        setNewChatCreated(true);
        toggle(false);
        setCurrChatroom(response.data);
    };
      
  
    return (
        <Wrapper $isOpen={isOpen}>
            <div className="flex items-center justify-between">
                <IoIosArrowRoundBack 
                  className="cursor-pointer" 
                  size={40} 
                  onClick={() => { toggle(false); }}
                />
                <h3 className="text-2xl">New Chat</h3>
                <IoMdCheckmark 
                  className={Object.keys(selectedUser).length === 0 ? 'text-neutral-400 cursor-default' : 'text-neutral-50 cursor-pointer'} 
                  size={30} 
                  onClick={startChatReq}
                />
            </div>
            <Form {...form}>
                <form onSubmit={form.handleSubmit(searchUserReq)} className="flex py-5 ">
                    <FormField
                        control={form.control}
                        name="usernameSearch"
                        render={({ field }) => (
                            <FormItem className="w-full">
                                <FormControl>
                                    <Input {...field} className="flex-1 text-neutral-600 bg-neutral-100" ref={searchInputRef} placeholder="Search name"/>
                                </FormControl>
                            </FormItem>
                        )}
                    />
                </form>
            </Form>
            <div className="mt-[5px]">
                {usernames == null || usernames.length == 0 ? '' : usernames.map((username) => (
                    <Button 
                      variant={selectedUser?._id == username._id ? 'selected' : 'inverse'} 
                      className="w-full my-[10px] text-base" 
                      key={username._id} 
                      onClick={() => { setSelectedUser(username); }}
                    >
                      {username.username}
                    </Button>
                ))}
            </div>
        </Wrapper>
    );
};
