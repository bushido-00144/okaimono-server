class User {
    constructor(userId) {
        this.userId = userId;
    }

    getUserInfo() {
        return fetch('/user/'+this.userId).then((res)=>{
            return res.json();
        }).then((userData)=>{
            this.userName = userData.UserName;
            return this.getUserRemainder();
        }).then((remainderData)=>{
            this.remainder = remainderData.Remainder;
        });
    }

    getUserRemainder() {
        return fetch('/remainder/'+this.userId).then((data)=>{
            return data.json();
        }).then((remainderData)=>{
            return remainderData;
        });
    }
}


// DOM Loaded
document.addEventListener('DOMContentLoaded', ()=>{
    let user = new User(1);
    let promise = new Promise((resolve, reject)=>{
        resolve(user.getUserInfo());
    }).then(()=>{
        document.getElementById('user-name').innerHTML = user.userName;
        document.getElementById('user-remainder').innerHTML = user.remainder;
        document.getElementById('userprof-button').addEventListener('click', ()=>{
        });
    });
})
