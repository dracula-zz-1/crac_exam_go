export namespace config {
	
	export class ExamConfig {
	    total: number;
	    single: number;
	    multiple: number;
	    time_minutes: number;
	    pass_score: number;
	
	    static createFrom(source: any = {}) {
	        return new ExamConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.single = source["single"];
	        this.multiple = source["multiple"];
	        this.time_minutes = source["time_minutes"];
	        this.pass_score = source["pass_score"];
	    }
	}

}

export namespace models {
	
	export class ExamQuestionDetail {
	    id: number;
	    exam_id: number;
	    question_id: number;
	    question_text: string;
	    option_a: string;
	    option_b: string;
	    option_c: string;
	    option_d: string;
	    correct_answer: string;
	    user_answer: string;
	    is_correct: boolean;
	    type: number;
	    image_data: string;
	
	    static createFrom(source: any = {}) {
	        return new ExamQuestionDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.exam_id = source["exam_id"];
	        this.question_id = source["question_id"];
	        this.question_text = source["question_text"];
	        this.option_a = source["option_a"];
	        this.option_b = source["option_b"];
	        this.option_c = source["option_c"];
	        this.option_d = source["option_d"];
	        this.correct_answer = source["correct_answer"];
	        this.user_answer = source["user_answer"];
	        this.is_correct = source["is_correct"];
	        this.type = source["type"];
	        this.image_data = source["image_data"];
	    }
	}
	export class ExamRecord {
	    id: number;
	    category: string;
	    exam_date: time.Time;
	    duration_seconds: number;
	    user_id: number;
	    score: number;
	    total_questions: number;
	    correct_count: number;
	
	    static createFrom(source: any = {}) {
	        return new ExamRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.category = source["category"];
	        this.exam_date = this.convertValues(source["exam_date"], time.Time);
	        this.duration_seconds = source["duration_seconds"];
	        this.user_id = source["user_id"];
	        this.score = source["score"];
	        this.total_questions = source["total_questions"];
	        this.correct_count = source["correct_count"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ExamStatisticsData {
	    id: number;
	    category: string;
	    exam_date: time.Time;
	    total_questions: number;
	    correct_questions: number;
	    pass_rate: number;
	    duration_seconds: number;
	    score: number;
	
	    static createFrom(source: any = {}) {
	        return new ExamStatisticsData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.category = source["category"];
	        this.exam_date = this.convertValues(source["exam_date"], time.Time);
	        this.total_questions = source["total_questions"];
	        this.correct_questions = source["correct_questions"];
	        this.pass_rate = source["pass_rate"];
	        this.duration_seconds = source["duration_seconds"];
	        this.score = source["score"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Question {
	    id: number;
	    J: string;
	    P: string;
	    I: string;
	    Q: string;
	    T: string;
	    A: string;
	    B: string;
	    C: string;
	    D: string;
	    F: string;
	    LA: number;
	    LB: number;
	    LC: number;
	    type: number;
	    user_id: number;
	    type_text: string;
	    has_image: string;
	
	    static createFrom(source: any = {}) {
	        return new Question(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.J = source["J"];
	        this.P = source["P"];
	        this.I = source["I"];
	        this.Q = source["Q"];
	        this.T = source["T"];
	        this.A = source["A"];
	        this.B = source["B"];
	        this.C = source["C"];
	        this.D = source["D"];
	        this.F = source["F"];
	        this.LA = source["LA"];
	        this.LB = source["LB"];
	        this.LC = source["LC"];
	        this.type = source["type"];
	        this.user_id = source["user_id"];
	        this.type_text = source["type_text"];
	        this.has_image = source["has_image"];
	    }
	}
	export class User {
	    id: number;
	    username: string;
	    id_card: string;
	    last_login: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new User(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.username = source["username"];
	        this.id_card = source["id_card"];
	        this.last_login = this.convertValues(source["last_login"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace services {
	
	export class AppInfo {
	    name: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	    }
	}
	export class ExamResult {
	    exam_id: number;
	    category: string;
	    exam_date: time.Time;
	    duration_seconds: number;
	    score: number;
	    correct_count: number;
	    total_count: number;
	    pass_exam: boolean;
	    pass_score: number;
	
	    static createFrom(source: any = {}) {
	        return new ExamResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.exam_id = source["exam_id"];
	        this.category = source["category"];
	        this.exam_date = this.convertValues(source["exam_date"], time.Time);
	        this.duration_seconds = source["duration_seconds"];
	        this.score = source["score"];
	        this.correct_count = source["correct_count"];
	        this.total_count = source["total_count"];
	        this.pass_exam = source["pass_exam"];
	        this.pass_score = source["pass_score"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ExamStartResponse {
	    exam_id: number;
	    questions: models.Question[];
	    config: config.ExamConfig;
	
	    static createFrom(source: any = {}) {
	        return new ExamStartResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.exam_id = source["exam_id"];
	        this.questions = this.convertValues(source["questions"], models.Question);
	        this.config = this.convertValues(source["config"], config.ExamConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ExamStatisticsResult {
	    max_score: number;
	    latest_score: number;
	    latest_duration: number;
	    avg_pass_rate: number;
	    total_exams: number;
	
	    static createFrom(source: any = {}) {
	        return new ExamStatisticsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.max_score = source["max_score"];
	        this.latest_score = source["latest_score"];
	        this.latest_duration = source["latest_duration"];
	        this.avg_pass_rate = source["avg_pass_rate"];
	        this.total_exams = source["total_exams"];
	    }
	}
	export class ImportResult {
	    success: boolean;
	    message: string;
	    imported_count: number;
	    total_count: number;
	    stats: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new ImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.imported_count = source["imported_count"];
	        this.total_count = source["total_count"];
	        this.stats = source["stats"];
	    }
	}
	export class PageDataResult {
	    data: models.Question[];
	    total: number;
	    page: number;
	    page_size: number;
	    total_pages: number;
	
	    static createFrom(source: any = {}) {
	        return new PageDataResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], models.Question);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.page_size = source["page_size"];
	        this.total_pages = source["total_pages"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UserLoginResponse {
	    success: boolean;
	    message?: string;
	    user_info?: models.User;
	
	    static createFrom(source: any = {}) {
	        return new UserLoginResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.user_info = this.convertValues(source["user_info"], models.User);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UserStatisticsResult {
	    total_exams: number;
	    total_practices: number;
	    total_errors: number;
	    total_favorites: number;
	    avg_exam_score: number;
	    avg_practice_rate: number;
	    exam_pass_rate: number;
	    last_exam_date?: time.Time;
	    last_practice_date?: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new UserStatisticsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_exams = source["total_exams"];
	        this.total_practices = source["total_practices"];
	        this.total_errors = source["total_errors"];
	        this.total_favorites = source["total_favorites"];
	        this.avg_exam_score = source["avg_exam_score"];
	        this.avg_practice_rate = source["avg_practice_rate"];
	        this.exam_pass_rate = source["exam_pass_rate"];
	        this.last_exam_date = this.convertValues(source["last_exam_date"], time.Time);
	        this.last_practice_date = this.convertValues(source["last_practice_date"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace time {
	
	export class Time {
	
	
	    static createFrom(source: any = {}) {
	        return new Time(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

